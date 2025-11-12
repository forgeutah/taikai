package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/forgeutah/taikai/internal/auth"
	"github.com/forgeutah/taikai/pkg/jwt"
)

// AuthMiddleware creates middleware that validates JWT tokens
type AuthMiddleware struct {
	jwtManager *jwt.Manager
	// blacklist will be used to check if tokens are revoked
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtManager *jwt.Manager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// Authenticate is middleware that requires a valid JWT token
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		// Check for Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondError(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				respondError(w, http.StatusUnauthorized, "token has expired")
			} else {
				respondError(w, http.StatusUnauthorized, "invalid token")
			}
			return
		}

		// Add user info to context
		ctx := auth.SetUserContext(r.Context(), claims.UserID, claims.Email, claims.Name)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthenticate is middleware that validates JWT if present but doesn't require it
func (m *AuthMiddleware) OptionalAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				claims, err := m.jwtManager.ValidateToken(parts[1])
				if err == nil {
					// Add user info to context if token is valid
					ctx := auth.SetUserContext(r.Context(), claims.UserID, claims.Email, claims.Name)
					r = r.WithContext(ctx)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

// respondError sends a JSON error response
func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":{"message":"` + message + `"}}`))
}
