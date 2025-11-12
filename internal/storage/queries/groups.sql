-- name: GetGroup :one
SELECT * FROM groups WHERE id = $1 LIMIT 1;

-- name: GetGroupBySlug :one
SELECT * FROM groups WHERE slug = $1 LIMIT 1;

-- name: ListGroupsByOrganization :many
SELECT * FROM groups WHERE organization_id = $1 ORDER BY created_at DESC;

-- name: ListAllGroups :many
SELECT * FROM groups ORDER BY created_at DESC;

-- name: CreateGroup :one
INSERT INTO groups (
    organization_id, name, slug, description, logo_url,
    primary_location, community_chat_url, community_chat_name
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpdateGroup :one
UPDATE groups
SET name = $2, description = $3, logo_url = $4, primary_location = $5,
    community_chat_url = $6, community_chat_name = $7, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGroup :exec
DELETE FROM groups WHERE id = $1;
