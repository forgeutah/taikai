-- Venue Queries

-- name: GetVenue :one
SELECT id, name, address, city, state, postal_code, country,
       directions, capacity, timezone, is_virtual, created_at
FROM venues
WHERE id = $1;

-- name: ListVenues :many
SELECT id, name, address, city, state, postal_code, country,
       directions, capacity, timezone, is_virtual, created_at
FROM venues
WHERE
    CASE
        WHEN $1::text != '' THEN
            name ILIKE '%' || $1 || '%' OR
            city ILIKE '%' || $1 || '%' OR
            state ILIKE '%' || $1 || '%'
        ELSE TRUE
    END
ORDER BY name ASC
LIMIT $2 OFFSET $3;

-- name: CountVenues :one
SELECT COUNT(*)
FROM venues
WHERE
    CASE
        WHEN $1::text != '' THEN
            name ILIKE '%' || $1 || '%' OR
            city ILIKE '%' || $1 || '%' OR
            state ILIKE '%' || $1 || '%'
        ELSE TRUE
    END;

-- name: CreateVenue :one
INSERT INTO venues (
    name, address, city, state, postal_code, country,
    directions, capacity, timezone, is_virtual
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING id, name, address, city, state, postal_code, country,
          directions, capacity, timezone, is_virtual, created_at;

-- name: UpdateVenue :one
UPDATE venues
SET
    name = COALESCE($2, name),
    address = COALESCE($3, address),
    city = COALESCE($4, city),
    state = COALESCE($5, state),
    postal_code = COALESCE($6, postal_code),
    country = COALESCE($7, country),
    directions = COALESCE($8, directions),
    capacity = COALESCE($9, capacity),
    timezone = COALESCE($10, timezone),
    is_virtual = COALESCE($11, is_virtual)
WHERE id = $1
RETURNING id, name, address, city, state, postal_code, country,
          directions, capacity, timezone, is_virtual, created_at;

-- name: DeleteVenue :exec
DELETE FROM venues WHERE id = $1;
