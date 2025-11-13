-- +goose Up
-- Event Sponsors
CREATE TABLE event_sponsors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    company_name VARCHAR(255) NOT NULL,
    logo_url VARCHAR(512),
    description TEXT,
    email VARCHAR(255) NOT NULL,
    website VARCHAR(512),
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_event_sponsors_event_id ON event_sponsors(event_id);

-- +goose Down
DROP TABLE IF EXISTS event_sponsors;
