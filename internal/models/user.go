package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email             string             `bson:"email" json:"email"`
	PasswordHash      string             `bson:"passwordHash" json:"-"`
	SMTPHost          string             `bson:"smtpHost,omitempty" json:"smtpHost,omitempty"`
	SMTPPort          int                `bson:"smtpPort,omitempty" json:"smtpPort,omitempty"`
	SMTPUser          string             `bson:"smtpUser,omitempty" json:"smtpUser,omitempty"`
	SMTPPassEncrypted string             `bson:"smtpPassEncrypted,omitempty" json:"-"`
	FromEmail         string             `bson:"fromEmail,omitempty" json:"fromEmail,omitempty"`
	UseTLS            bool               `bson:"useTLS" json:"useTLS"`
	APIKey            string             `bson:"apiKey" json:"apiKey"`
	EmailLogTTLDays   int                `bson:"emailLogTTLDays" json:"emailLogTTLDays"` // 0 = never delete
	CreatedAt         time.Time          `bson:"createdAt" json:"createdAt"`
}

// RegisterRequest represents registration request payload
type RegisterRequest struct {
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=6"`
	SMTPHost     string `json:"smtpHost,omitempty"`
	SMTPPort     int    `json:"smtpPort,omitempty"`
	SMTPUser     string `json:"smtpUser,omitempty"`
	SMTPPassword string `json:"smtpPassword,omitempty"`
	FromEmail    string `json:"fromEmail,omitempty"`
	UseTLS       *bool  `json:"useTLS,omitempty"`
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token  string `json:"token"`
	APIKey string `json:"apiKey"`
	User   User   `json:"user"`
}

// UpdateTTLRequest represents request to update email log TTL
type UpdateTTLRequest struct {
	EmailLogTTLDays int `json:"emailLogTTLDays" validate:"gte=0"`
}
