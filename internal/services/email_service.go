package services

import (
	"context"
	"crypto/tls"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/gomail.v2"

	"send-jer/config"
	"send-jer/internal/database"
	"send-jer/internal/models"
	"send-jer/internal/utils"
)

var emailLogsCollection = "email_logs"

// EmailService handles email-related operations
type EmailService struct{}

// NewEmailService creates a new EmailService
func NewEmailService() *EmailService {
	return &EmailService{}
}

// SendEmail sends an email using the user's SMTP or default SMTP
func (s *EmailService) SendEmail(user *models.User, to, subject, body string, isHTML bool) error {
	// Determine which SMTP to use
	smtpHost := user.SMTPHost
	smtpPort := user.SMTPPort
	smtpUser := user.SMTPUser
	smtpPass := ""
	fromEmail := user.FromEmail
	useTLS := user.UseTLS

	// Use default SMTP if user hasn't configured their own
	if smtpHost == "" {
		smtpHost = config.AppConfig.DefaultSMTPHost
		smtpPort = config.AppConfig.DefaultSMTPPort
		smtpUser = config.AppConfig.DefaultSMTPUser
		smtpPass = config.AppConfig.DefaultSMTPPass
		fromEmail = config.AppConfig.DefaultFromEmail
		useTLS = config.AppConfig.DefaultUseTLS
	} else {
		// Decrypt user's SMTP password
		if user.SMTPPassEncrypted != "" {
			decrypted, err := utils.Decrypt(user.SMTPPassEncrypted)
			if err != nil {
				return err
			}
			smtpPass = decrypted
		}
	}

	// Validate SMTP configuration
	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return errors.New("SMTP configuration is incomplete")
	}

	// Create email message
	m := gomail.NewMessage()
	m.SetHeader("From", fromEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	
	if isHTML {
		m.SetBody("text/html", body)
	} else {
		m.SetBody("text/plain", body)
	}

	// Create dialer
	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	
	// Configure TLS
	if useTLS {
		d.TLSConfig = &tls.Config{ServerName: smtpHost}
	} else {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Send email
	return d.DialAndSend(m)
}

// LogEmail creates a log entry for a sent email
func (s *EmailService) LogEmail(userID primitive.ObjectID, to, subject, body, status, errorMessage string) (*models.EmailLog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log := models.EmailLog{
		UserID:       userID,
		To:           to,
		Subject:      subject,
		Body:         body,
		Status:       status,
		ErrorMessage: errorMessage,
		SentAt:       time.Now(),
	}

	collection := database.Database.Collection(emailLogsCollection)
	result, err := collection.InsertOne(ctx, log)
	if err != nil {
		return nil, err
	}

	log.ID = result.InsertedID.(primitive.ObjectID)
	return &log, nil
}

// GetEmailLogs retrieves email logs for a user
func (s *EmailService) GetEmailLogs(userID string, limit int64) ([]models.EmailLog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	collection := database.Database.Collection(emailLogsCollection)
	
	opts := options.Find().SetSort(bson.D{{Key: "sentAt", Value: -1}})
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := collection.Find(ctx, bson.M{"userId": objID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []models.EmailLog
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

// CleanupOldLogs deletes old email logs based on user's TTL setting
func (s *EmailService) CleanupOldLogs(userID primitive.ObjectID, ttlDays int) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cutoff := time.Now().AddDate(0, 0, -ttlDays)
	
	collection := database.Database.Collection(emailLogsCollection)
	result, err := collection.DeleteMany(ctx, bson.M{
		"userId": userID,
		"sentAt": bson.M{"$lt": cutoff},
	})
	
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}
