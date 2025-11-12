-- name: GetGroup :one
SELECT * FROM groups WHERE id = $1 LIMIT 1;

-- name: GetGroupBySlug :one
SELECT * FROM groups WHERE slug = $1 LIMIT 1;

-- name: ListGroupsByOrganization :many
SELECT * FROM groups
WHERE organization_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAllGroups :many
SELECT * FROM groups
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountGroupsByOrganization :one
SELECT COUNT(*) FROM groups WHERE organization_id = $1;

-- name: CountAllGroups :one
SELECT COUNT(*) FROM groups;

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

-- Group Admin Queries

-- name: IsGroupAdmin :one
SELECT EXISTS(
    SELECT 1 FROM group_admins
    WHERE user_id = $1 AND group_id = $2
) AS is_admin;

-- name: GetGroupAdmins :many
SELECT u.id, u.email, u.name, u.avatar_url, ga.created_at as assigned_at
FROM group_admins ga
JOIN users u ON u.id = ga.user_id
WHERE ga.group_id = $1 AND u.is_active = true
ORDER BY ga.created_at ASC;

-- name: AddGroupAdmin :one
INSERT INTO group_admins (user_id, group_id)
VALUES ($1, $2)
ON CONFLICT (user_id, group_id) DO NOTHING
RETURNING *;

-- name: RemoveGroupAdmin :exec
DELETE FROM group_admins
WHERE user_id = $1 AND group_id = $2;

-- name: GetUserGroupAdminRoles :many
SELECT group_id FROM group_admins
WHERE user_id = $1;

-- name: GetGroupWithOrganization :one
SELECT
    g.*,
    o.name as org_name,
    o.community_chat_url as org_chat_url,
    o.community_chat_name as org_chat_name
FROM groups g
JOIN organizations o ON g.organization_id = o.id
WHERE g.id = $1;
