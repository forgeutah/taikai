package auth

import (
	"context"
	"database/sql"
	"time"
)

// PostgresBlacklist handles token blacklisting using PostgreSQL
type PostgresBlacklist struct {
	db *sql.DB
}

// NewPostgresBlacklist creates a new PostgreSQL-based blacklist
func NewPostgresBlacklist(db *sql.DB) *PostgresBlacklist {
	return &PostgresBlacklist{
		db: db,
	}
}

// BlacklistToken adds a token to the blacklist
func (p *PostgresBlacklist) BlacklistToken(ctx context.Context, token string, expiration time.Duration) error {
	expiresAt := time.Now().Add(expiration)

	query := `
		INSERT INTO token_blacklist (token, expires_at)
		VALUES ($1, $2)
		ON CONFLICT (token) DO UPDATE SET expires_at = EXCLUDED.expires_at
	`

	_, err := p.db.ExecContext(ctx, query, token, expiresAt)
	return err
}

// IsBlacklisted checks if a token is blacklisted
func (p *PostgresBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM token_blacklist
			WHERE token = $1 AND expires_at > NOW()
		)
	`

	var exists bool
	err := p.db.QueryRowContext(ctx, query, token).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// CleanupExpired removes expired tokens from the blacklist
// This should be called periodically by a background job or trigger
func (p *PostgresBlacklist) CleanupExpired(ctx context.Context) error {
	query := `DELETE FROM token_blacklist WHERE expires_at <= NOW()`
	_, err := p.db.ExecContext(ctx, query)
	return err
}
