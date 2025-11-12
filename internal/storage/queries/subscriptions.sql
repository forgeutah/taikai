-- Organization Subscriptions

-- name: CreateOrgSubscription :one
INSERT INTO org_subscriptions (user_id, organization_id, notify_email, notify_sms, notify_discord)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, organization_id)
DO UPDATE SET
    notify_email = EXCLUDED.notify_email,
    notify_sms = EXCLUDED.notify_sms,
    notify_discord = EXCLUDED.notify_discord
RETURNING id, user_id, organization_id, notify_email, notify_sms, notify_discord, created_at;

-- name: DeleteOrgSubscription :exec
DELETE FROM org_subscriptions
WHERE user_id = $1 AND organization_id = $2;

-- name: GetOrgSubscription :one
SELECT id, user_id, organization_id, notify_email, notify_sms, notify_discord, created_at
FROM org_subscriptions
WHERE user_id = $1 AND organization_id = $2;

-- name: UpdateOrgSubscription :one
UPDATE org_subscriptions
SET
    notify_email = COALESCE($3, notify_email),
    notify_sms = COALESCE($4, notify_sms),
    notify_discord = COALESCE($5, notify_discord)
WHERE user_id = $1 AND organization_id = $2
RETURNING id, user_id, organization_id, notify_email, notify_sms, notify_discord, created_at;

-- name: GetOrgSubscribers :many
SELECT
    os.id, os.user_id, os.organization_id,
    os.notify_email, os.notify_sms, os.notify_discord, os.created_at,
    u.email, u.name, u.avatar_url
FROM org_subscriptions os
JOIN users u ON os.user_id = u.id
WHERE os.organization_id = $1
ORDER BY os.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountOrgSubscribers :one
SELECT COUNT(*)
FROM org_subscriptions
WHERE organization_id = $1;

-- Group Subscriptions

-- name: CreateGroupSubscription :one
INSERT INTO group_subscriptions (user_id, group_id, notify_email, notify_sms, notify_discord)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, group_id)
DO UPDATE SET
    notify_email = EXCLUDED.notify_email,
    notify_sms = EXCLUDED.notify_sms,
    notify_discord = EXCLUDED.notify_discord
RETURNING id, user_id, group_id, notify_email, notify_sms, notify_discord, created_at;

-- name: DeleteGroupSubscription :exec
DELETE FROM group_subscriptions
WHERE user_id = $1 AND group_id = $2;

-- name: GetGroupSubscription :one
SELECT id, user_id, group_id, notify_email, notify_sms, notify_discord, created_at
FROM group_subscriptions
WHERE user_id = $1 AND group_id = $2;

-- name: UpdateGroupSubscription :one
UPDATE group_subscriptions
SET
    notify_email = COALESCE($3, notify_email),
    notify_sms = COALESCE($4, notify_sms),
    notify_discord = COALESCE($5, notify_discord)
WHERE user_id = $1 AND group_id = $2
RETURNING id, user_id, group_id, notify_email, notify_sms, notify_discord, created_at;

-- name: GetGroupSubscribers :many
SELECT
    gs.id, gs.user_id, gs.group_id,
    gs.notify_email, gs.notify_sms, gs.notify_discord, gs.created_at,
    u.email, u.name, u.avatar_url
FROM group_subscriptions gs
JOIN users u ON gs.user_id = u.id
WHERE gs.group_id = $1
ORDER BY gs.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountGroupSubscribers :one
SELECT COUNT(*)
FROM group_subscriptions
WHERE group_id = $1;

-- User's Subscriptions

-- name: GetUserOrgSubscriptions :many
SELECT
    os.id, os.user_id, os.organization_id,
    os.notify_email, os.notify_sms, os.notify_discord, os.created_at,
    o.name as organization_name, o.slug as organization_slug, o.logo_url
FROM org_subscriptions os
JOIN organizations o ON os.organization_id = o.id
WHERE os.user_id = $1
ORDER BY o.name ASC;

-- name: GetUserGroupSubscriptions :many
SELECT
    gs.id, gs.user_id, gs.group_id,
    gs.notify_email, gs.notify_sms, gs.notify_discord, gs.created_at,
    g.name as group_name, g.slug as group_slug, g.organization_id
FROM group_subscriptions gs
JOIN groups g ON gs.group_id = g.id
WHERE gs.user_id = $1
ORDER BY g.name ASC;

-- Check if user has any subscription (org or group) for an event

-- name: CheckUserHasOrgSubscription :one
SELECT EXISTS(
    SELECT 1 FROM org_subscriptions
    WHERE user_id = $1 AND organization_id = $2
);

-- name: CheckUserHasGroupSubscription :one
SELECT EXISTS(
    SELECT 1 FROM group_subscriptions
    WHERE user_id = $1 AND group_id = $2
);

-- Get all subscribers for an event (deduplicated)
-- This query gets org subscribers first, then group subscribers who aren't already org subscribers

-- name: GetEventSubscribers :many
WITH org_subs AS (
    SELECT DISTINCT
        os.user_id,
        os.notify_email,
        os.notify_sms,
        os.notify_discord,
        u.email,
        u.name
    FROM org_subscriptions os
    JOIN users u ON os.user_id = u.id
    JOIN events e ON e.group_id = (SELECT group_id FROM events WHERE id = $1 LIMIT 1)
    JOIN groups g ON g.id = e.group_id
    WHERE os.organization_id = g.organization_id
),
group_subs AS (
    SELECT DISTINCT
        gs.user_id,
        gs.notify_email,
        gs.notify_sms,
        gs.notify_discord,
        u.email,
        u.name
    FROM group_subscriptions gs
    JOIN users u ON gs.user_id = u.id
    JOIN events e ON e.group_id = $1
    WHERE gs.group_id = e.group_id
    AND gs.user_id NOT IN (SELECT user_id FROM org_subs)
)
SELECT user_id, notify_email, notify_sms, notify_discord, email, name
FROM org_subs
UNION ALL
SELECT user_id, notify_email, notify_sms, notify_discord, email, name
FROM group_subs;

-- Get email subscribers for an event (filter by notify_email = true)

-- name: GetEventEmailSubscribers :many
WITH org_subs AS (
    SELECT DISTINCT
        os.user_id,
        u.email,
        u.name
    FROM org_subscriptions os
    JOIN users u ON os.user_id = u.id
    JOIN events e ON e.id = $1
    JOIN groups g ON g.id = e.group_id
    WHERE os.organization_id = g.organization_id
    AND os.notify_email = true
),
group_subs AS (
    SELECT DISTINCT
        gs.user_id,
        u.email,
        u.name
    FROM group_subscriptions gs
    JOIN users u ON gs.user_id = u.id
    JOIN events e ON e.id = $1
    WHERE gs.group_id = e.group_id
    AND gs.notify_email = true
    AND gs.user_id NOT IN (SELECT user_id FROM org_subs)
)
SELECT user_id, email, name
FROM org_subs
UNION ALL
SELECT user_id, email, name
FROM group_subs;
