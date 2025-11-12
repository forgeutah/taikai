-- +goose Up
-- Organization Subscriptions
CREATE TABLE org_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    notify_email BOOLEAN DEFAULT TRUE,
    notify_sms BOOLEAN DEFAULT FALSE,
    notify_discord BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, organization_id)
);

CREATE INDEX idx_org_subscriptions_user_id ON org_subscriptions(user_id);
CREATE INDEX idx_org_subscriptions_organization_id ON org_subscriptions(organization_id);

-- Group Subscriptions
CREATE TABLE group_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    notify_email BOOLEAN DEFAULT TRUE,
    notify_sms BOOLEAN DEFAULT FALSE,
    notify_discord BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, group_id)
);

CREATE INDEX idx_group_subscriptions_user_id ON group_subscriptions(user_id);
CREATE INDEX idx_group_subscriptions_group_id ON group_subscriptions(group_id);

-- +goose Down
DROP TABLE IF EXISTS group_subscriptions;
DROP TABLE IF EXISTS org_subscriptions;
