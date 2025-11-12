package jwt

import (
	"testing"
	"time"
)

func TestGenerateAndValidateToken(t *testing.T) {
	manager := NewManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)

	userID := "test-user-id"
	email := "test@example.com"
	name := "Test User"
	roles := []string{"user"}

	// Generate token pair
	tokens, err := manager.GenerateTokenPair(userID, email, name, roles)
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	if tokens.AccessToken == "" {
		t.Error("Access token is empty")
	}
	if tokens.RefreshToken == "" {
		t.Error("Refresh token is empty")
	}
	if tokens.ExpiresIn <= 0 {
		t.Error("ExpiresIn should be positive")
	}

	// Validate access token
	claims, err := manager.ValidateToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}
	if claims.Name != name {
		t.Errorf("Expected name %s, got %s", name, claims.Name)
	}
}

func TestInvalidToken(t *testing.T) {
	manager := NewManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)

	_, err := manager.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

func TestExpiredToken(t *testing.T) {
	manager := NewManager("test-secret-key", -1*time.Second, 7*24*time.Hour)

	tokens, err := manager.GenerateTokenPair("user-id", "test@example.com", "Test", nil)
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	// Token should be expired immediately
	time.Sleep(2 * time.Second)

	_, err = manager.ValidateToken(tokens.AccessToken)
	if err != ErrExpiredToken {
		t.Errorf("Expected ErrExpiredToken, got %v", err)
	}
}

func TestRefreshAccessToken(t *testing.T) {
	manager := NewManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)

	tokens, err := manager.GenerateTokenPair("user-id", "test@example.com", "Test", nil)
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	// Refresh access token
	newAccessToken, err := manager.RefreshAccessToken(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}

	if newAccessToken == "" {
		t.Error("New access token is empty")
	}

	// Validate new token
	claims, err := manager.ValidateToken(newAccessToken)
	if err != nil {
		t.Fatalf("Failed to validate new access token: %v", err)
	}

	if claims.UserID != "user-id" {
		t.Errorf("Expected user ID 'user-id', got %s", claims.UserID)
	}
}
