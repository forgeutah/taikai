package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"os"
	"time"

	"github.com/forgeutah/taikai/internal/auth"
	"github.com/forgeutah/taikai/pkg/email"
	"github.com/forgeutah/taikai/pkg/jwt"
	"github.com/google/uuid"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	db           *sql.DB
	jwtManager   *jwt.Manager
	emailService *email.Service
	blacklist    *auth.RedisBlacklist
	baseURL      string
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(db *sql.DB, jwtManager *jwt.Manager, emailService *email.Service, blacklist *auth.RedisBlacklist) *AuthHandler {
	return &AuthHandler{
		db:           db,
		jwtManager:   jwtManager,
		emailService: emailService,
		blacklist:    blacklist,
		baseURL:      os.Getenv("BASE_URL"),
	}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Timezone string `json:"timezone"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}

// UserResponse represents a user in responses
type UserResponse struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	AvatarURL     *string   `json:"avatar_url"`
	Timezone      string    `json:"timezone"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Validate email
	if _, err := mail.ParseAddress(req.Email); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeValidation, "Invalid email address")
		return
	}

	// Validate password
	if err := auth.ValidatePasswordStrength(req.Password); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeValidation, err.Error())
		return
	}

	// Validate name
	if req.Name == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeValidation, "Name is required")
		return
	}

	// Default timezone if not provided
	if req.Timezone == "" {
		req.Timezone = "America/Denver"
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		RespondError(w, http.StatusInternalServer, ErrCodeInternalServer, "Failed to process registration")
		return
	}

	// Create user
	userID := uuid.New().String()
	query := `
		INSERT INTO users (id, email, password_hash, name, timezone, email_verified, is_active)
		VALUES ($1, $2, $3, $4, $5, false, true)
		RETURNING id, email, name, avatar_url, timezone, email_verified, created_at
	`

	var user UserResponse
	err = h.db.QueryRow(query, userID, req.Email, hashedPassword, req.Name, req.Timezone).Scan(
		&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.Timezone, &user.EmailVerified, &user.CreatedAt,
	)
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			RespondError(w, http.StatusConflict, ErrCodeConflict, "Email already registered")
			return
		}
		RespondError(w, http.StatusInternalServer, ErrCodeInternalServer, "Failed to create user")
		return
	}

	// Generate verification token
	token, expiresAt, err := auth.GenerateEmailVerificationToken()
	if err != nil {
		RespondError(w, http.StatusInternalServer, ErrCodeInternalServer, "Failed to generate verification token")
		return
	}

	// Store verification token
	_, err = h.db.Exec(`
		INSERT INTO email_verification_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`, user.ID, token, expiresAt)
	if err != nil {
		RespondError(w, http.StatusInternalServer, ErrCodeInternalServer, "Failed to create verification token")
		return
	}

	// Send verification email
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", h.baseURL, token)
	emailData := map[string]string{
		"Name":            user.Name,
		"VerificationURL": verificationURL,
	}

	err = h.emailService.SendFromTemplate(user.Email, email.EmailVerificationTemplate, emailData)
	if err != nil {
		// Log error but don't fail registration
		fmt.Printf("Failed to send verification email: %v\n", err)
	}

	RespondSuccess(w, http.StatusCreated, map[string]interface{}{
		"user":    user,
		"message": "Registration successful. Please check your email to verify your account.",
	})
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Get user by email
	var user struct {
		ID            string
		Email         string
		PasswordHash  string
		Name          string
		AvatarURL     *string
		Timezone      string
		EmailVerified bool
		IsActive      bool
		CreatedAt     time.Time
	}

	query := `
		SELECT id, email, password_hash, name, avatar_url, timezone, email_verified, is_active, created_at
		FROM users
		WHERE email = $1 AND is_active = true
	`
	err := h.db.QueryRow(query, req.Email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.AvatarURL,
		&user.Timezone, &user.EmailVerified, &user.IsActive, &user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Invalid email or password")
			return
		}
		RespondError(w, http.StatusInternalServer, ErrCodeInternalServer, "Login failed")
		return
	}

	// Verify password
	if err := auth.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Invalid email or password")
		return
	}

	// Update last login
	_, err = h.db.Exec("UPDATE users SET last_login_at = NOW() WHERE id = $1", user.ID)
	if err != nil {
		// Log but don't fail login
		fmt.Printf("Failed to update last login: %v\n", err)
	}

	// Generate tokens
	tokens, err := h.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Name, nil)
	if err != nil {
		RespondError(w, http.StatusInternalServer, ErrCodeInternalServer, "Failed to generate tokens")
		return
	}

	response := AuthResponse{
		User: UserResponse{
			ID:            user.ID,
			Email:         user.Email,
			Name:          user.Name,
			AvatarURL:     user.AvatarURL,
			Timezone:      user.Timezone,
			EmailVerified: user.EmailVerified,
			CreatedAt:     user.CreatedAt,
		},
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}

	RespondSuccess(w, http.StatusOK, response)
}

// Logout handles user logout (blacklists refresh token)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Validate and extract expiration from token
	claims, err := h.jwtManager.ValidateToken(req.RefreshToken)
	if err != nil {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Invalid refresh token")
		return
	}

	// Calculate remaining time until expiration
	expiration := time.Until(claims.ExpiresAt.Time)
	if expiration > 0 {
		// Blacklist the token
		err = h.blacklist.BlacklistToken(r.Context(), req.RefreshToken, expiration)
		if err != nil {
			// Log but don't fail logout
			fmt.Printf("Failed to blacklist token: %v\n", err)
		}
	}

	RespondSuccess(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// Refresh handles token refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Check if token is blacklisted
	if h.blacklist != nil {
		blacklisted, err := h.blacklist.IsBlacklisted(r.Context(), req.RefreshToken)
		if err != nil {
			RespondError(w, http.StatusInternalServer, ErrCodeInternalServer, "Token validation failed")
			return
		}
		if blacklisted {
			RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Token has been revoked")
			return
		}
	}

	// Generate new access token
	newAccessToken, err := h.jwtManager.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		if err == jwt.ErrExpiredToken {
			RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Refresh token has expired")
		} else {
			RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Invalid refresh token")
		}
		return
	}

	// Get token expiration duration from JWT manager
	accessTokenDuration := int64(15 * 60) // 15 minutes in seconds

	RespondSuccess(w, http.StatusOK, map[string]interface{}{
		"access_token": newAccessToken,
		"expires_in":   accessTokenDuration,
	})
}

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Verification token is required")
		return
	}

	// Get token from database
	var userID string
	var expiresAt time.Time
	query := `
		SELECT user_id, expires_at
		FROM email_verification_tokens
		WHERE token = $1
	`
	err := h.db.QueryRow(query, token).Scan(&userID, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid or expired verification token")
			return
		}
		RespondError(w, http.StatusInternalServer, ErrCodeInternalServer, "Verification failed")
		return
	}

	// Check if token is expired
	if time.Now().After(expiresAt) {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Verification token has expired")
		return
	}

	// Update user email_verified status
	_, err = h.db.Exec("UPDATE users SET email_verified = true, updated_at = NOW() WHERE id = $1", userID)
	if err != nil {
		RespondError(w, http.StatusInternalServer, ErrCodeInternalServer, "Failed to verify email")
		return
	}

	// Delete the verification token
	_, err = h.db.Exec("DELETE FROM email_verification_tokens WHERE token = $1", token)
	if err != nil {
		// Log but don't fail
		fmt.Printf("Failed to delete verification token: %v\n", err)
	}

	// Get user info for welcome email
	var user struct {
		Email string
		Name  string
	}
	err = h.db.QueryRow("SELECT email, name FROM users WHERE id = $1", userID).Scan(&user.Email, &user.Name)
	if err == nil {
		// Send welcome email
		emailData := map[string]string{
			"Name":    user.Name,
			"BaseURL": h.baseURL,
		}
		h.emailService.SendFromTemplate(user.Email, email.WelcomeTemplate, emailData)
	}

	RespondSuccess(w, http.StatusOK, map[string]string{"message": "Email verified successfully"})
}

// ForgotPassword handles forgot password requests
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Get user
	var user struct {
		ID   string
		Name string
	}
	err := h.db.QueryRow("SELECT id, name FROM users WHERE email = $1 AND is_active = true", req.Email).Scan(&user.ID, &user.Name)
	if err != nil {
		// Return success even if email not found (security best practice)
		RespondSuccess(w, http.StatusOK, map[string]string{"message": "If the email exists, a password reset link has been sent"})
		return
	}

	// Generate reset token
	token, expiresAt, err := auth.GeneratePasswordResetToken()
	if err != nil {
		RespondError(w, http.StatusInternalServer, ErrCodeInternalServer, "Failed to generate reset token")
		return
	}

	// Delete old reset tokens for this user
	_, err = h.db.Exec("DELETE FROM password_reset_tokens WHERE user_id = $1", user.ID)
	if err != nil {
		fmt.Printf("Failed to delete old reset tokens: %v\n", err)
	}

	// Store reset token
	_, err = h.db.Exec(`
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`, user.ID, token, expiresAt)
	if err != nil {
		RespondError(w, http.StatusInternalServer, ErrCodeInternalServer, "Failed to create reset token")
		return
	}

	// Send reset email
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", h.baseURL, token)
	emailData := map[string]string{
		"Name":     user.Name,
		"ResetURL": resetURL,
	}

	err = h.emailService.SendFromTemplate(req.Email, email.PasswordResetTemplate, emailData)
	if err != nil {
		fmt.Printf("Failed to send reset email: %v\n", err)
	}

	RespondSuccess(w, http.StatusOK, map[string]string{"message": "If the email exists, a password reset link has been sent"})
}

// ResetPassword handles password reset
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Validate new password
	if err := auth.ValidatePasswordStrength(req.Password); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeValidation, err.Error())
		return
	}

	// Get token from database
	var userID string
	var expiresAt time.Time
	query := `
		SELECT user_id, expires_at
		FROM password_reset_tokens
		WHERE token = $1
	`
	err := h.db.QueryRow(query, req.Token).Scan(&userID, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid or expired reset token")
			return
		}
		RespondError(w, http.StatusInternalServer, ErrCodeInternalServer, "Password reset failed")
		return
	}

	// Check if token is expired
	if time.Now().After(expiresAt) {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Reset token has expired")
		return
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		RespondError(w, http.StatusInternalServer, ErrCodeInternalServer, "Failed to process password")
		return
	}

	// Update password
	_, err = h.db.Exec("UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2", hashedPassword, userID)
	if err != nil {
		RespondError(w, http.StatusInternalServer, ErrCodeInternalServer, "Failed to update password")
		return
	}

	// Delete the reset token
	_, err = h.db.Exec("DELETE FROM password_reset_tokens WHERE token = $1", req.Token)
	if err != nil {
		fmt.Printf("Failed to delete reset token: %v\n", err)
	}

	RespondSuccess(w, http.StatusOK, map[string]string{"message": "Password reset successfully"})
}
