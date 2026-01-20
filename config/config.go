package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port string

	// MongoDB
	MongoURI      string
	MongoDatabase string

	// JWT
	JWTSecret string

	// Encryption
	EncryptionKey string

	// Default SMTP
	DefaultSMTPHost     string
	DefaultSMTPPort     int
	DefaultSMTPUser     string
	DefaultSMTPPass     string
	DefaultFromEmail    string
	DefaultUseTLS       bool

	// Worker
	WorkerConcurrency int

	// Cron
	CleanupCronSchedule string
}

var AppConfig *Config

func LoadConfig() {
	// Load .env file if exists
	_ = godotenv.Load()

	AppConfig = &Config{
		Port:                getEnv("PORT", "8080"),
		MongoURI:            getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDatabase:       getEnv("MONGODB_DATABASE", "email_webhook_service"),
		JWTSecret:           getEnv("JWT_SECRET", "change-this-secret"),
		EncryptionKey:       getEnv("ENCRYPTION_KEY", "change-this-32-byte-key-please"),
		DefaultSMTPHost:     getEnv("DEFAULT_SMTP_HOST", "smtp.gmail.com"),
		DefaultSMTPPort:     getEnvAsInt("DEFAULT_SMTP_PORT", 587),
		DefaultSMTPUser:     getEnv("DEFAULT_SMTP_USER", ""),
		DefaultSMTPPass:     getEnv("DEFAULT_SMTP_PASS", ""),
		DefaultFromEmail:    getEnv("DEFAULT_FROM_EMAIL", ""),
		DefaultUseTLS:       getEnvAsBool("DEFAULT_USE_TLS", true),
		WorkerConcurrency:   getEnvAsInt("WORKER_CONCURRENCY", 5),
		CleanupCronSchedule: getEnv("CLEANUP_CRON_SCHEDULE", "0 0 * * *"),
	}

	log.Println("Configuration loaded successfully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
