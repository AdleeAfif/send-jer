package api

import (
	"encoding/json"
	"net/http"

	"send-jer/internal/api/middleware"
	"send-jer/internal/models"
	"send-jer/internal/services"
)

// UserHandler handles user-related endpoints
type UserHandler struct {
	userService  *services.UserService
	emailService *services.EmailService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService:  services.NewUserService(),
		emailService: services.NewEmailService(),
	}
}

// UpdateEmailTTL updates the email log retention days
func (h *UserHandler) UpdateEmailTTL(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	var req models.UpdateTTLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update TTL
	err := h.userService.UpdateEmailLogTTL(userID, req.EmailLogTTLDays)
	if err != nil {
		http.Error(w, "Failed to update TTL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Email log TTL updated successfully",
	})
}

// GetEmailLogs retrieves email logs for the authenticated user
func (h *UserHandler) GetEmailLogs(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	// Get logs (limit to 100 most recent)
	logs, err := h.emailService.GetEmailLogs(userID, 100)
	if err != nil {
		http.Error(w, "Failed to retrieve logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"count":   len(logs),
		"logs":    logs,
	})
}
