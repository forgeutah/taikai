package auth

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisBlacklist handles token blacklisting using Redis
type RedisBlacklist struct {
	client *redis.Client
}

// NewRedisBlacklist creates a new Redis blacklist
func NewRedisBlacklist(client *redis.Client) *RedisBlacklist {
	return &RedisBlacklist{
		client: client,
	}
}

// BlacklistToken adds a token to the blacklist
func (r *RedisBlacklist) BlacklistToken(ctx context.Context, token string, expiration time.Duration) error {
	key := "blacklist:" + token
	return r.client.Set(ctx, key, "1", expiration).Err()
}

// IsBlacklisted checks if a token is blacklisted
func (r *RedisBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := "blacklist:" + token
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val != "", nil
}
