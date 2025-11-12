-- +goose Up
CREATE TABLE venues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100),
    directions TEXT,
    capacity INTEGER,
    timezone VARCHAR(100) NOT NULL, -- IANA timezone
    is_virtual BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_venues_city_state ON venues(city, state);

-- +goose Down
DROP TABLE IF EXISTS venues;
