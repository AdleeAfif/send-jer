package database

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"send-jer/config"
)

var (
	Client   *mongo.Client
	Database *mongo.Database
)

// Connect establishes connection to MongoDB
func Connect() error {
	timeout := time.Duration(config.AppConfig.MongoConnectTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(config.AppConfig.MongoURI).
		SetConnectTimeout(timeout).
		SetServerSelectionTimeout(timeout)

	log.Printf("Connecting to MongoDB database %q via hosts: %s", config.AppConfig.MongoDatabase, strings.Join(clientOptions.Hosts, ", "))

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return wrapMongoError("connect", err, clientOptions.Hosts)
	}

	// Ping the database
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		_ = client.Disconnect(context.Background())
		return wrapMongoError("ping", err, clientOptions.Hosts)
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

func wrapMongoError(operation string, err error, hosts []string) error {
	message := fmt.Sprintf("mongo %s failed for hosts [%s]: %v", operation, strings.Join(hosts, ", "), err)
	errText := err.Error()

	if strings.Contains(errText, "ReplicaSetNoPrimary") {
		message += ". Atlas has no reachable primary from this app. Common causes are a paused cluster, an IP allow-list/firewall problem, or a VPN/proxy/TLS inspection layer breaking the handshake to some replica-set members"
	}

	if strings.Contains(errText, "tls: internal error") {
		message += ". The TLS handshake is being rejected by the remote side; verify Atlas network access for this machine, remove any TLS-inspecting proxy/VPN, and confirm the cluster is healthy"
	}

	return fmt.Errorf("%s", message)
}
