package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"send-jer/internal/api/middleware"
	"send-jer/internal/models"
	"send-jer/internal/services"
)

// AdminHandler handles admin-only user management endpoints.
type AdminHandler struct {
	userService *services.UserService
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler() *AdminHandler {
	return &AdminHandler{
		userService: services.NewUserService(),
	}
}

// ListUsers returns all users for admin management.
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.ListUsers()
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"count":   len(users),
		"users":   users,
	})
}

// GetUser returns a single managed user.
func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["id"]

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user":    user,
	})
}

// UpdateUser allows an admin to update another user's management fields.
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	currentUser := r.Context().Value(middleware.UserKey).(*models.User)
	userID := mux.Vars(r)["id"]

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if currentUser.ID.Hex() == userID && req.Role == models.RoleUser {
		http.Error(w, "Admin cannot downgrade their own role", http.StatusBadRequest)
		return
	}

	user, err := h.userService.UpdateManagedUser(userID, req)
	if err != nil {
		switch err.Error() {
		case "invalid user ID", "invalid role", "emailLogTTLDays must be greater than or equal to 0", "no valid fields provided for update":
			http.Error(w, err.Error(), http.StatusBadRequest)
		case "user not found":
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User updated successfully",
		"user":    user,
	})
}

// DeleteUser removes a managed user account.
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	currentUser := r.Context().Value(middleware.UserKey).(*models.User)
	userID := mux.Vars(r)["id"]

	if currentUser.ID.Hex() == userID {
		http.Error(w, "Admin cannot delete their own account", http.StatusBadRequest)
		return
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		switch err.Error() {
		case "invalid user ID":
			http.Error(w, err.Error(), http.StatusBadRequest)
		case "user not found":
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User deleted successfully",
	})
}
