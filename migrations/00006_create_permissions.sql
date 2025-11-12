-- +goose Up
-- Organization Admins
CREATE TABLE org_admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, organization_id)
);

CREATE INDEX idx_org_admins_user_id ON org_admins(user_id);
CREATE INDEX idx_org_admins_organization_id ON org_admins(organization_id);

-- Group Admins
CREATE TABLE group_admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, group_id)
);

CREATE INDEX idx_group_admins_user_id ON group_admins(user_id);
CREATE INDEX idx_group_admins_group_id ON group_admins(group_id);

-- Event Hosts
CREATE TABLE event_hosts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, event_id)
);

CREATE INDEX idx_event_hosts_user_id ON event_hosts(user_id);
CREATE INDEX idx_event_hosts_event_id ON event_hosts(event_id);

-- +goose Down
DROP TABLE IF EXISTS event_hosts;
DROP TABLE IF EXISTS group_admins;
DROP TABLE IF EXISTS org_admins;
