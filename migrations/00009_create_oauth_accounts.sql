-- +goose Up
CREATE TYPE oauth_provider AS ENUM ('google', 'github');

CREATE TABLE oauth_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider oauth_provider NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT, -- Should be encrypted at application level
    refresh_token TEXT, -- Should be encrypted at application level
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(provider, provider_user_id)
);

CREATE INDEX idx_oauth_accounts_user_id ON oauth_accounts(user_id);

-- +goose Down
DROP TABLE IF EXISTS oauth_accounts;
DROP TYPE IF EXISTS oauth_provider;
