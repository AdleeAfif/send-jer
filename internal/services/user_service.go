package services

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"send-jer/internal/database"
	"send-jer/internal/models"
	"send-jer/internal/utils"
)

var usersCollection = "users"

// UserService handles user-related operations
type UserService struct{}

// NewUserService creates a new UserService
func NewUserService() *UserService {
	return &UserService{}
}

func normalizeUserRole(user *models.User) {
	if user != nil && user.Role == "" {
		user.Role = models.RoleUser
	}
}

// Register creates a new user
func (s *UserService) Register(req models.RegisterRequest) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user already exists
	collection := database.Database.Collection(usersCollection)
	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		return nil, errors.New("user already exists")
	}

	userRole := models.RoleUser
	userCount, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if userCount == 0 {
		userRole = models.RoleAdmin
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Generate API key
	apiKey, err := utils.GenerateAPIKey()
	if err != nil {
		return nil, err
	}

	// Prepare user object
	user := models.User{
		Email:           req.Email,
		Role:            userRole,
		PasswordHash:    passwordHash,
		APIKey:          apiKey,
		EmailLogTTLDays: 30, // default 30 days
		UseTLS:          true,
		CreatedAt:       time.Now(),
	}

	// Handle optional SMTP configuration
	if req.SMTPHost != "" {
		user.SMTPHost = req.SMTPHost
		user.SMTPPort = req.SMTPPort
		user.SMTPUser = req.SMTPUser

		// Encrypt SMTP password if provided
		if req.SMTPPassword != "" {
			encrypted, err := utils.Encrypt(req.SMTPPassword)
			if err != nil {
				return nil, err
			}
			user.SMTPPassEncrypted = encrypted
		}

		user.FromEmail = req.FromEmail
		if req.UseTLS != nil {
			user.UseTLS = *req.UseTLS
		}
	}

	// Insert user
	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return &user, nil
}

// Login authenticates a user
func (s *UserService) Login(req models.LoginRequest) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := database.Database.Collection(usersCollection)
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	normalizeUserRole(&user)
	return &user, nil
}

// GetUserByAPIKey retrieves a user by their API key
func (s *UserService) GetUserByAPIKey(apiKey string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := database.Database.Collection(usersCollection)
	var user models.User
	err := collection.FindOne(ctx, bson.M{"apiKey": apiKey}).Decode(&user)
	if err != nil {
		return nil, errors.New("invalid API key")
	}

	normalizeUserRole(&user)
	return &user, nil
}

// GetUserByID retrieves a user by their ID
func (s *UserService) GetUserByID(userID string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	collection := database.Database.Collection(usersCollection)
	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return nil, errors.New("user not found")
	}

	normalizeUserRole(&user)
	return &user, nil
}

// UpdateEmailLogTTL updates the email log retention days for a user
func (s *UserService) UpdateEmailLogTTL(userID string, ttlDays int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	collection := database.Database.Collection(usersCollection)
	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"emailLogTTLDays": ttlDays}},
	)

	return err
}

// ListUsers retrieves all users ordered by creation time.
func (s *UserService) ListUsers() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := database.Database.Collection(usersCollection)
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	for i := range users {
		normalizeUserRole(&users[i])
	}

	return users, nil
}

// DeleteUser removes a user account.
func (s *UserService) DeleteUser(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	collection := database.Database.Collection(usersCollection)
	result, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// UpdateManagedUser applies admin-managed changes to a user.
func (s *UserService) UpdateManagedUser(userID string, req models.UpdateUserRequest) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	updates := bson.M{}
	if req.Role != "" {
		if req.Role != models.RoleAdmin && req.Role != models.RoleUser {
			return nil, errors.New("invalid role")
		}
		updates["role"] = req.Role
	}
	if req.EmailLogTTLDays != nil {
		if *req.EmailLogTTLDays < 0 {
			return nil, errors.New("emailLogTTLDays must be greater than or equal to 0")
		}
		updates["emailLogTTLDays"] = *req.EmailLogTTLDays
	}
	if len(updates) == 0 {
		return nil, errors.New("no valid fields provided for update")
	}

	collection := database.Database.Collection(usersCollection)
	result := collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": updates},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var user models.User
	if err := result.Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	normalizeUserRole(&user)
	return &user, nil
}

// GetAllUsersWithTTL retrieves all users with TTL > 0 for cleanup job
func (s *UserService) GetAllUsersWithTTL() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := database.Database.Collection(usersCollection)
	cursor, err := collection.Find(ctx, bson.M{"emailLogTTLDays": bson.M{"$gt": 0}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}
