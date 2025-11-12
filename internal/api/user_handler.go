package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/forgeutah/taikai/internal/auth"
)

// UserHandler handles user-related requests
type UserHandler struct {
	db *sql.DB
}

// NewUserHandler creates a new user handler
func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{
		db: db,
	}
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Name      *string `json:"name,omitempty"`
	Bio       *string `json:"bio,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Timezone  *string `json:"timezone,omitempty"`
	DiscordID *string `json:"discord_id,omitempty"`
}

// GetMe returns the current user's profile
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Not authenticated")
		return
	}

	var user UserResponse
	query := `
		SELECT id, email, name, avatar_url, timezone, email_verified, created_at
		FROM users
		WHERE id = $1 AND is_active = true
	`
	err := h.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.Timezone, &user.EmailVerified, &user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondError(w, http.StatusNotFound, ErrCodeNotFound, "User not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get user")
		return
	}

	RespondSuccess(w, http.StatusOK, user)
}

// UpdateMe updates the current user's profile
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Not authenticated")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Build update query dynamically
	query := "UPDATE users SET updated_at = NOW()"
	args := []interface{}{userID}
	argIndex := 2

	if req.Name != nil {
		query += ", name = $" + string(rune(argIndex+'0'))
		args = append(args, *req.Name)
		argIndex++
	}
	if req.Bio != nil {
		query += ", bio = $" + string(rune(argIndex+'0'))
		args = append(args, *req.Bio)
		argIndex++
	}
	if req.Phone != nil {
		query += ", phone = $" + string(rune(argIndex+'0'))
		args = append(args, *req.Phone)
		argIndex++
	}
	if req.Timezone != nil {
		query += ", timezone = $" + string(rune(argIndex+'0'))
		args = append(args, *req.Timezone)
		argIndex++
	}
	if req.DiscordID != nil {
		query += ", discord_id = $" + string(rune(argIndex+'0'))
		args = append(args, *req.DiscordID)
		argIndex++
	}

	query += " WHERE id = $1 AND is_active = true RETURNING id, email, name, avatar_url, timezone, email_verified, created_at"

	var user UserResponse
	err := h.db.QueryRow(query, args...).Scan(
		&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.Timezone, &user.EmailVerified, &user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondError(w, http.StatusNotFound, ErrCodeNotFound, "User not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to update user")
		return
	}

	RespondSuccess(w, http.StatusOK, user)
}

// DeleteMe soft deletes the current user's account
func (h *UserHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Not authenticated")
		return
	}

	// Soft delete user
	_, err := h.db.Exec("UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1", userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to delete account")
		return
	}

	RespondSuccess(w, http.StatusOK, map[string]string{"message": "Account deleted successfully"})
}

// GetUserByID returns a public user profile
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	// This will be implemented when we add Chi router to extract URL params
	RespondError(w, http.StatusNotImplemented, ErrCodeInternalServer, "Not implemented yet")
}
