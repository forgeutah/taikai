package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantError bool
	}{
		{
			name:      "valid password",
			password:  "ValidPass123!",
			wantError: false,
		},
		{
			name:      "password too short",
			password:  "short",
			wantError: true,
		},
		{
			name:      "minimum length password",
			password:  "12345678",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if tt.wantError {
				if err == nil {
					t.Errorf("HashPassword() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("HashPassword() unexpected error: %v", err)
				return
			}
			if hash == "" {
				t.Errorf("HashPassword() returned empty hash")
			}
			if hash == tt.password {
				t.Errorf("HashPassword() returned unhashed password")
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "TestPassword123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	tests := []struct {
		name         string
		hashedPass   string
		plainPass    string
		wantError    bool
		errorMessage string
	}{
		{
			name:       "correct password",
			hashedPass: hash,
			plainPass:  password,
			wantError:  false,
		},
		{
			name:         "incorrect password",
			hashedPass:   hash,
			plainPass:    "WrongPassword",
			wantError:    true,
			errorMessage: "invalid password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyPassword(tt.hashedPass, tt.plainPass)
			if tt.wantError {
				if err == nil {
					t.Errorf("VerifyPassword() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("VerifyPassword() unexpected error: %v", err)
			}
		})
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantError bool
	}{
		{
			name:      "valid password",
			password:  "ValidPass123!",
			wantError: false,
		},
		{
			name:      "minimum length",
			password:  "12345678",
			wantError: false,
		},
		{
			name:      "too short",
			password:  "1234567",
			wantError: true,
		},
		{
			name:      "empty password",
			password:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if tt.wantError {
				if err == nil {
					t.Errorf("ValidatePasswordStrength() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ValidatePasswordStrength() unexpected error: %v", err)
			}
		})
	}
}
