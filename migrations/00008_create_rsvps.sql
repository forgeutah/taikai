-- +goose Up
CREATE TYPE rsvp_status AS ENUM ('attending', 'not_attending');

CREATE TABLE event_rsvps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    status rsvp_status NOT NULL DEFAULT 'attending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, event_id)
);

CREATE INDEX idx_event_rsvps_user_id ON event_rsvps(user_id);
CREATE INDEX idx_event_rsvps_event_id ON event_rsvps(event_id);
CREATE INDEX idx_event_rsvps_status ON event_rsvps(status);

-- +goose Down
DROP TABLE IF EXISTS event_rsvps;
DROP TYPE IF EXISTS rsvp_status;
