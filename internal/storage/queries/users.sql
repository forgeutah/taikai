-- name: GetUser :one
SELECT * FROM users WHERE id = $1 AND is_active = true LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 AND is_active = true LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
    email, password_hash, name, avatar_url, bio, phone, timezone, discord_id, email_verified
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET name = $2, avatar_url = $3, bio = $4, phone = $5,
    timezone = $6, discord_id = $7, updated_at = NOW()
WHERE id = $1 AND is_active = true
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1 AND is_active = true;

-- name: VerifyUserEmail :exec
UPDATE users
SET email_verified = true, updated_at = NOW()
WHERE id = $1;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW()
WHERE id = $1;

-- name: SoftDeleteUser :exec
UPDATE users
SET is_active = false, updated_at = NOW()
WHERE id = $1;
