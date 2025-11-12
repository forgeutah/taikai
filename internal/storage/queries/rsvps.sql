-- RSVP Queries

-- name: CreateOrUpdateRSVP :one
INSERT INTO event_rsvps (user_id, event_id, status)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, event_id)
DO UPDATE SET
    status = EXCLUDED.status,
    updated_at = NOW()
RETURNING id, user_id, event_id, status, created_at, updated_at;

-- name: GetUserRSVP :one
SELECT id, user_id, event_id, status, created_at, updated_at
FROM event_rsvps
WHERE user_id = $1 AND event_id = $2;

-- name: DeleteRSVP :exec
DELETE FROM event_rsvps
WHERE user_id = $1 AND event_id = $2;

-- name: GetEventRSVPCount :one
SELECT COUNT(*)
FROM event_rsvps
WHERE event_id = $1 AND status = 'attending';

-- name: GetEventRSVPs :many
SELECT r.id, r.user_id, r.event_id, r.status, r.created_at, r.updated_at,
       u.email, u.name, u.avatar_url
FROM event_rsvps r
JOIN users u ON r.user_id = u.id
WHERE r.event_id = $1
ORDER BY r.created_at DESC;

-- name: GetEventRSVPsByStatus :many
SELECT r.id, r.user_id, r.event_id, r.status, r.created_at, r.updated_at,
       u.email, u.name, u.avatar_url
FROM event_rsvps r
JOIN users u ON r.user_id = u.id
WHERE r.event_id = $1 AND r.status = $2
ORDER BY r.created_at DESC;

-- name: GetUserRSVPs :many
SELECT r.id, r.user_id, r.event_id, r.status, r.created_at, r.updated_at,
       e.title, e.slug, e.description, e.start_time, e.end_time, e.timezone,
       e.venue_id, e.capacity, e.status as event_status
FROM event_rsvps r
JOIN events e ON r.event_id = e.id
WHERE r.user_id = $1 AND r.status = 'attending' AND e.start_time > NOW()
ORDER BY e.start_time ASC
LIMIT $2 OFFSET $3;

-- name: CountUserRSVPs :one
SELECT COUNT(*)
FROM event_rsvps r
JOIN events e ON r.event_id = e.id
WHERE r.user_id = $1 AND r.status = 'attending' AND e.start_time > NOW();

-- name: GetUserHostedEvents :many
SELECT DISTINCT e.id, e.group_id, e.title, e.slug, e.description, e.status,
       e.start_time, e.end_time, e.timezone, e.venue_id, e.capacity,
       e.youtube_url, e.created_by, e.created_at, e.updated_at,
       (SELECT COUNT(*) FROM event_rsvps WHERE event_id = e.id AND status = 'attending') as rsvp_count
FROM events e
JOIN event_hosts eh ON e.id = eh.event_id
WHERE eh.user_id = $1 AND e.start_time > NOW()
ORDER BY e.start_time ASC
LIMIT $2 OFFSET $3;

-- name: CountUserHostedEvents :one
SELECT COUNT(DISTINCT e.id)
FROM events e
JOIN event_hosts eh ON e.id = eh.event_id
WHERE eh.user_id = $1 AND e.start_time > NOW();
