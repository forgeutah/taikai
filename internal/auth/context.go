package auth

import (
	"context"
)

// ContextKey is the type for context keys
type ContextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// UserEmailKey is the context key for user email
	UserEmailKey ContextKey = "user_email"
	// UserNameKey is the context key for user name
	UserNameKey ContextKey = "user_name"
)

// GetUserIDFromContext retrieves the user ID from the context
func GetUserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(UserIDKey).(string)
	return userID
}

// GetUserEmailFromContext retrieves the user email from the context
func GetUserEmailFromContext(ctx context.Context) string {
	email, _ := ctx.Value(UserEmailKey).(string)
	return email
}

// GetUserNameFromContext retrieves the user name from the context
func GetUserNameFromContext(ctx context.Context) string {
	name, _ := ctx.Value(UserNameKey).(string)
	return name
}

// SetUserContext adds user information to the context
func SetUserContext(ctx context.Context, userID, email, name string) context.Context {
	ctx = context.WithValue(ctx, UserIDKey, userID)
	ctx = context.WithValue(ctx, UserEmailKey, email)
	ctx = context.WithValue(ctx, UserNameKey, name)
	return ctx
}
