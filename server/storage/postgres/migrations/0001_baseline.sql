-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

create table if not exists orgs (
	id                       uuid      default uuid_generate_v4() not null primary key,
	created_at               timestamp default now(),
	updated_at               timestamp default now(),
	org_name                 text not null unique
);

create table if not exists users (
	id                       uuid      default uuid_generate_v4() not null primary key,
	created_at               timestamp default now(),
	updated_at               timestamp default now(),
	username                text not null unique,
	first_name               text,
	last_name                text,
	city                     text,
	zip                      text,
	email                    text not null unique,
	job_title                text,
	org_id                   uuid references orgs(id)
);

-- +goose Down
drop table if exists users;
drop table if exists orgs;
