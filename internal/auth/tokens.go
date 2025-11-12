package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

const (
	// TokenLength is the length of verification/reset tokens
	TokenLength = 32
	// EmailVerificationTokenExpiry is how long email verification tokens are valid
	EmailVerificationTokenExpiry = 24 * time.Hour
	// PasswordResetTokenExpiry is how long password reset tokens are valid
	PasswordResetTokenExpiry = 1 * time.Hour
)

// GenerateRandomToken generates a cryptographically secure random token
func GenerateRandomToken() (string, error) {
	bytes := make([]byte, TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateEmailVerificationToken generates a token for email verification
func GenerateEmailVerificationToken() (token string, expiresAt time.Time, err error) {
	token, err = GenerateRandomToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt = time.Now().Add(EmailVerificationTokenExpiry)
	return token, expiresAt, nil
}

// GeneratePasswordResetToken generates a token for password reset
func GeneratePasswordResetToken() (token string, expiresAt time.Time, err error) {
	token, err = GenerateRandomToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt = time.Now().Add(PasswordResetTokenExpiry)
	return token, expiresAt, nil
}
