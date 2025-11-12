-- +goose Up
CREATE TYPE event_status AS ENUM ('draft', 'published', 'cancelled');

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT NOT NULL,
    status event_status DEFAULT 'draft',
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    timezone VARCHAR(100) NOT NULL, -- IANA timezone for display
    venue_id UUID REFERENCES venues(id) ON DELETE SET NULL,
    capacity INTEGER, -- Nullable for unlimited capacity
    youtube_url VARCHAR(512),

    -- Recurring event fields
    parent_event_id UUID REFERENCES events(id) ON DELETE CASCADE, -- NULL for non-recurring or series parent
    recurrence_rule TEXT, -- RRULE format (RFC 5545)
    recurrence_end_date TIMESTAMP WITH TIME ZONE, -- When recurrence stops
    recurrence_count INTEGER, -- Alternative to end_date: stop after N occurrences
    is_recurring_parent BOOLEAN DEFAULT FALSE, -- True for the template event

    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT valid_time_range CHECK (end_time > start_time)
);

CREATE INDEX idx_events_group_id ON events(group_id);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_start_time ON events(start_time);
CREATE INDEX idx_events_slug ON events(slug);
CREATE INDEX idx_events_created_by ON events(created_by);
CREATE INDEX idx_events_parent_event_id ON events(parent_event_id); -- For recurring series

-- Composite index for common query: upcoming published events for a group
CREATE INDEX idx_events_group_status_time ON events(group_id, status, start_time);

-- +goose Down
DROP TABLE IF EXISTS events;
DROP TYPE IF EXISTS event_status;
