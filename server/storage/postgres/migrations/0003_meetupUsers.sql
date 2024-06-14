-- +goose Up

ALTER TABLE users ADD COLUMN meetup_id int unique;
ALTER TABLE users ADD COLUMN group_id UUID[];

ALTER TABLE groups ADD COLUMN admins UUID[];

-- make sure that users only join groups that exist
ALTER TABLE users ADD CONSTRAINT fk_group_id FOREIGN KEY (group_id) REFERENCES groups(id);

-- make sure that groups only have admins that exist
ALTER TABLE groups ADD Constraint fk_org_id FOREIGN KEY (admins) REFERENCES users(id);

-- +goose Down
ALTER TABLE users DROP COLUMN meetup_id;
ALTER TABLE users DROP COLUMN group_id;
ALTER TABLE users DROP CONSTRAINT fk_group_id;
ALTER TABLE groups DROP COLUMN admins;

