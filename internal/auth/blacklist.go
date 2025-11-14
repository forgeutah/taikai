package auth

import (
	"context"
	"time"
)

// TokenBlacklist defines the interface for token blacklist implementations
type TokenBlacklist interface {
	BlacklistToken(ctx context.Context, token string, expiration time.Duration) error
	IsBlacklisted(ctx context.Context, token string) (bool, error)
}
