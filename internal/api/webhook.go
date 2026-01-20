package api

import (
	"encoding/json"
	"net/http"

	"send-jer/internal/api/middleware"
	"send-jer/internal/models"
	"send-jer/internal/workers"
)

// WebhookHandler handles webhook endpoints
type WebhookHandler struct {
	emailWorker *workers.EmailWorker
}

// NewWebhookHandler creates a new WebhookHandler
func NewWebhookHandler(emailWorker *workers.EmailWorker) *WebhookHandler {
	return &WebhookHandler{
		emailWorker: emailWorker,
	}
}

// SendEmail handles the webhook to send emails
func (h *WebhookHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by APIKeyAuth middleware)
	user := r.Context().Value(middleware.UserKey).(*models.User)

	var req models.SendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.To == "" || req.Subject == "" || req.Body == "" {
		http.Error(w, "To, subject, and body are required", http.StatusBadRequest)
		return
	}

	// Queue email for sending
	logID, err := h.emailWorker.QueueEmail(user, req)
	if err != nil {
		http.Error(w, "Failed to queue email", http.StatusInternalServerError)
		return
	}

	response := models.SendEmailResponse{
		Success: true,
		Message: "Email queued for sending",
		LogID:   logID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}
