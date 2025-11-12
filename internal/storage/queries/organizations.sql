-- name: GetOrganization :one
SELECT * FROM organizations WHERE id = $1 LIMIT 1;

-- name: GetOrganizationBySlug :one
SELECT * FROM organizations WHERE slug = $1 LIMIT 1;

-- name: ListOrganizations :many
SELECT * FROM organizations ORDER BY created_at DESC;

-- name: CreateOrganization :one
INSERT INTO organizations (
    name, slug, description, logo_url, website_url,
    community_chat_url, community_chat_name
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: UpdateOrganization :one
UPDATE organizations
SET name = $2, description = $3, logo_url = $4, website_url = $5,
    community_chat_url = $6, community_chat_name = $7, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteOrganization :exec
DELETE FROM organizations WHERE id = $1;

-- Organization Admin Queries

-- name: IsOrgAdmin :one
SELECT EXISTS(
    SELECT 1 FROM org_admins
    WHERE user_id = $1 AND organization_id = $2
) AS is_admin;

-- name: GetOrgAdmins :many
SELECT u.id, u.email, u.name, u.avatar_url, oa.created_at as assigned_at
FROM org_admins oa
JOIN users u ON u.id = oa.user_id
WHERE oa.organization_id = $1 AND u.is_active = true
ORDER BY oa.created_at ASC;

-- name: AddOrgAdmin :one
INSERT INTO org_admins (user_id, organization_id)
VALUES ($1, $2)
ON CONFLICT (user_id, organization_id) DO NOTHING
RETURNING *;

-- name: RemoveOrgAdmin :exec
DELETE FROM org_admins
WHERE user_id = $1 AND organization_id = $2;

-- name: GetUserOrgAdminRoles :many
SELECT organization_id FROM org_admins
WHERE user_id = $1;
