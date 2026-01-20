package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EmailLog represents an email log entry
type EmailLog struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"userId" json:"userId"`
	To           string             `bson:"to" json:"to"`
	Subject      string             `bson:"subject" json:"subject"`
	Body         string             `bson:"body" json:"body"`
	Status       string             `bson:"status" json:"status"` // "success" or "failed"
	ErrorMessage string             `bson:"errorMessage,omitempty" json:"errorMessage,omitempty"`
	SentAt       time.Time          `bson:"sentAt" json:"sentAt"`
}

// SendEmailRequest represents webhook request to send email
type SendEmailRequest struct {
	To      string `json:"to" validate:"required,email"`
	Subject string `json:"subject" validate:"required"`
	Body    string `json:"body" validate:"required"`
	IsHTML  bool   `json:"isHTML"`
}

// SendEmailResponse represents webhook response
type SendEmailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	LogID   string `json:"logId,omitempty"`
}
