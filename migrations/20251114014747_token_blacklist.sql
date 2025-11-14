-- +goose Up
-- +goose StatementBegin
CREATE TABLE token_blacklist (
    token TEXT PRIMARY KEY,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for efficient cleanup of expired tokens
CREATE INDEX idx_token_blacklist_expires_at ON token_blacklist(expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS token_blacklist;
-- +goose StatementEnd
