-- +goose Up

ALTER TABLE users ADD COLUMN if not exists meetup_id int unique;
ALTER TABLE users ADD COLUMN if not exists group_id UUID[];
ALTER TABLE users ADD COLUMN if not exists state text;
ALTER TABLE groups ADD COLUMN if not exists admins UUID[];

-- +goose Down
ALTER TABLE users DROP COLUMN if exists meetup_id;
ALTER TABLE users DROP COLUMN if exists group_id;
ALTER TABLE users DROP COLUMN if exists state;
ALTER TABLE groups DROP COLUMN if exists admins;

