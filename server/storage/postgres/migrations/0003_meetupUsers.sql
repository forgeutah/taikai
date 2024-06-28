-- +goose Up

ALTER TABLE users ADD COLUMN meetup_id int unique;
ALTER TABLE users ADD COLUMN group_id UUID[];

ALTER TABLE groups ADD COLUMN admins UUID[];

-- +goose Down
ALTER TABLE users DROP COLUMN meetup_id;
ALTER TABLE users DROP COLUMN group_id;

