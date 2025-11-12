-- Event Queries

-- name: GetEvent :one
SELECT e.id, e.group_id, e.title, e.slug, e.description, e.status,
       e.start_time, e.end_time, e.timezone, e.venue_id, e.capacity,
       e.youtube_url, e.parent_event_id, e.recurrence_rule,
       e.recurrence_end_date, e.recurrence_count, e.is_recurring_parent,
       e.created_by, e.created_at, e.updated_at
FROM events e
WHERE e.id = $1;

-- name: GetEventBySlug :one
SELECT e.id, e.group_id, e.title, e.slug, e.description, e.status,
       e.start_time, e.end_time, e.timezone, e.venue_id, e.capacity,
       e.youtube_url, e.parent_event_id, e.recurrence_rule,
       e.recurrence_end_date, e.recurrence_count, e.is_recurring_parent,
       e.created_by, e.created_at, e.updated_at
FROM events e
WHERE e.slug = $1;

-- name: ListEvents :many
SELECT e.id, e.group_id, e.title, e.slug, e.description, e.status,
       e.start_time, e.end_time, e.timezone, e.venue_id, e.capacity,
       e.youtube_url, e.parent_event_id, e.recurrence_rule,
       e.recurrence_end_date, e.recurrence_count, e.is_recurring_parent,
       e.created_by, e.created_at, e.updated_at
FROM events e
WHERE
    -- Filter by group if provided
    CASE
        WHEN $1::uuid IS NOT NULL THEN e.group_id = $1
        ELSE TRUE
    END
    -- Filter by status if provided
    AND CASE
        WHEN $2::text != '' THEN e.status::text = $2
        ELSE TRUE
    END
    -- Filter by start date (events after this date)
    AND CASE
        WHEN $3::timestamptz IS NOT NULL THEN e.start_time >= $3
        ELSE TRUE
    END
    -- Filter by end date (events before this date)
    AND CASE
        WHEN $4::timestamptz IS NOT NULL THEN e.start_time <= $4
        ELSE TRUE
    END
ORDER BY e.start_time ASC
LIMIT $5 OFFSET $6;

-- name: CountEvents :one
SELECT COUNT(*)
FROM events e
WHERE
    -- Filter by group if provided
    CASE
        WHEN $1::uuid IS NOT NULL THEN e.group_id = $1
        ELSE TRUE
    END
    -- Filter by status if provided
    AND CASE
        WHEN $2::text != '' THEN e.status::text = $2
        ELSE TRUE
    END
    -- Filter by start date (events after this date)
    AND CASE
        WHEN $3::timestamptz IS NOT NULL THEN e.start_time >= $3
        ELSE TRUE
    END
    -- Filter by end date (events before this date)
    AND CASE
        WHEN $4::timestamptz IS NOT NULL THEN e.start_time <= $4
        ELSE TRUE
    END;

-- name: CreateEvent :one
INSERT INTO events (
    group_id, title, slug, description, status, start_time, end_time,
    timezone, venue_id, capacity, youtube_url, parent_event_id,
    recurrence_rule, recurrence_end_date, recurrence_count,
    is_recurring_parent, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
)
RETURNING id, group_id, title, slug, description, status, start_time, end_time,
          timezone, venue_id, capacity, youtube_url, parent_event_id,
          recurrence_rule, recurrence_end_date, recurrence_count,
          is_recurring_parent, created_by, created_at, updated_at;

-- name: UpdateEvent :one
UPDATE events
SET
    title = COALESCE($2, title),
    description = COALESCE($3, description),
    status = COALESCE($4, status),
    start_time = COALESCE($5, start_time),
    end_time = COALESCE($6, end_time),
    timezone = COALESCE($7, timezone),
    venue_id = COALESCE($8, venue_id),
    capacity = COALESCE($9, capacity),
    youtube_url = COALESCE($10, youtube_url),
    updated_at = NOW()
WHERE id = $1
RETURNING id, group_id, title, slug, description, status, start_time, end_time,
          timezone, venue_id, capacity, youtube_url, parent_event_id,
          recurrence_rule, recurrence_end_date, recurrence_count,
          is_recurring_parent, created_by, created_at, updated_at;

-- name: DeleteEvent :exec
DELETE FROM events WHERE id = $1;

-- Event Hosts Queries

-- name: GetEventHosts :many
SELECT eh.id, eh.user_id, eh.event_id, eh.created_at,
       u.email, u.name, u.avatar_url
FROM event_hosts eh
JOIN users u ON eh.user_id = u.id
WHERE eh.event_id = $1
ORDER BY eh.created_at ASC;

-- name: IsEventHost :one
SELECT EXISTS(
    SELECT 1 FROM event_hosts
    WHERE user_id = $1 AND event_id = $2
);

-- name: AddEventHost :one
INSERT INTO event_hosts (user_id, event_id)
VALUES ($1, $2)
ON CONFLICT (user_id, event_id) DO NOTHING
RETURNING id, user_id, event_id, created_at;

-- name: RemoveEventHost :exec
DELETE FROM event_hosts
WHERE user_id = $1 AND event_id = $2;

-- Event Speakers Queries

-- name: GetEventSpeakers :many
SELECT id, event_id, name, bio, avatar_url, display_order, created_at
FROM event_speakers
WHERE event_id = $1
ORDER BY display_order ASC, created_at ASC;

-- name: GetSpeaker :one
SELECT id, event_id, name, bio, avatar_url, display_order, created_at
FROM event_speakers
WHERE id = $1;

-- name: CreateSpeaker :one
INSERT INTO event_speakers (event_id, name, bio, avatar_url, display_order)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, event_id, name, bio, avatar_url, display_order, created_at;

-- name: UpdateSpeaker :one
UPDATE event_speakers
SET
    name = COALESCE($2, name),
    bio = COALESCE($3, bio),
    avatar_url = COALESCE($4, avatar_url),
    display_order = COALESCE($5, display_order)
WHERE id = $1
RETURNING id, event_id, name, bio, avatar_url, display_order, created_at;

-- name: DeleteSpeaker :exec
DELETE FROM event_speakers WHERE id = $1;

-- Event Schedule Queries

-- name: GetEventSchedules :many
SELECT id, event_id, time_offset_minutes, title, description, display_order, created_at
FROM event_schedules
WHERE event_id = $1
ORDER BY time_offset_minutes ASC, display_order ASC;

-- name: GetScheduleItem :one
SELECT id, event_id, time_offset_minutes, title, description, display_order, created_at
FROM event_schedules
WHERE id = $1;

-- name: CreateScheduleItem :one
INSERT INTO event_schedules (event_id, time_offset_minutes, title, description, display_order)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, event_id, time_offset_minutes, title, description, display_order, created_at;

-- name: UpdateScheduleItem :one
UPDATE event_schedules
SET
    time_offset_minutes = COALESCE($2, time_offset_minutes),
    title = COALESCE($3, title),
    description = COALESCE($4, description),
    display_order = COALESCE($5, display_order)
WHERE id = $1
RETURNING id, event_id, time_offset_minutes, title, description, display_order, created_at;

-- name: DeleteScheduleItem :exec
DELETE FROM event_schedules WHERE id = $1;
