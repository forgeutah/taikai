-- name: CreateEmailVerificationToken :one
INSERT INTO email_verification_tokens (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetEmailVerificationToken :one
SELECT * FROM email_verification_tokens
WHERE token = $1 AND expires_at > NOW()
LIMIT 1;

-- name: DeleteEmailVerificationToken :exec
DELETE FROM email_verification_tokens
WHERE token = $1;

-- name: DeleteUserEmailVerificationTokens :exec
DELETE FROM email_verification_tokens
WHERE user_id = $1;

-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPasswordResetToken :one
SELECT * FROM password_reset_tokens
WHERE token = $1 AND expires_at > NOW()
LIMIT 1;

-- name: DeletePasswordResetToken :exec
DELETE FROM password_reset_tokens
WHERE token = $1;

-- name: DeleteUserPasswordResetTokens :exec
DELETE FROM password_reset_tokens
WHERE user_id = $1;
