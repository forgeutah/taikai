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
