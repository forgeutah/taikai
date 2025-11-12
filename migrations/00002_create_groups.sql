-- +goose Up
CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    logo_url VARCHAR(512),
    primary_location TEXT,
    community_chat_url VARCHAR(512), -- Nullable, overrides org if set
    community_chat_name VARCHAR(100), -- Nullable
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_groups_organization_id ON groups(organization_id);
CREATE INDEX idx_groups_slug ON groups(slug);

-- +goose Down
DROP TABLE IF EXISTS groups;
