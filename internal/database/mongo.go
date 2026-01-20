package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"send-jer/config"
)

var (
	Client   *mongo.Client
	Database *mongo.Database
)

// Connect establishes connection to MongoDB
func Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.AppConfig.MongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	Client = client
	Database = client.Database(config.AppConfig.MongoDatabase)
	
	log.Println("Connected to MongoDB successfully")
	
	// Create indexes
	createIndexes()
	
	return nil
}

// Disconnect closes the MongoDB connection
func Disconnect() error {
	if Client == nil {
		return nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	return Client.Disconnect(ctx)
}

// createIndexes creates necessary database indexes
func createIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Users collection indexes
	usersCollection := Database.Collection("users")
	usersCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"email": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    map[string]interface{}{"apiKey": 1},
			Options: options.Index().SetUnique(true),
		},
	})

	// Email logs collection indexes
	emailLogsCollection := Database.Collection("email_logs")
	emailLogsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"userId": 1},
		},
		{
			Keys: map[string]interface{}{"sentAt": 1},
		},
		{
			Keys: map[string]interface{}{"userId": 1, "sentAt": -1},
		},
	})

	log.Println("Database indexes created successfully")
}
