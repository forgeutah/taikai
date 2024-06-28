-- +gooseUp
CREATE TABLE IF NOT EXISTS groups (
	id                       UUID DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
	created_at               TIMESTAMP DEFAULT now(),
	updated_at               TIMESTAMP DEFAULT now(),
	name                     TEXT NOT NULL,
	description              TEXT,
	org_id                   UUID REFERENCES orgs(id),
	meetup_id                TEXT
);

-- +gooseDown	
DROP TABLE IF EXISTS groups;
