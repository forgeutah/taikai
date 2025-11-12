-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255), -- Nullable for OAuth-only users
    name VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(512),
    bio TEXT,
    phone VARCHAR(20),
    timezone VARCHAR(100) NOT NULL DEFAULT 'America/Denver', -- IANA timezone
    discord_id VARCHAR(100),
    email_verified BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE, -- Soft delete flag
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_is_active ON users(is_active);

-- +goose Down
DROP TABLE IF EXISTS users;
