-- +goose Up
CREATE
EXTENSION IF NOT EXISTS "uuid-ossp";

create table hellos
(
    id                       uuid      default uuid_generate_v4() not null primary key,
    created_at               timestamp default now(),
    updated_at               timestamp default now(),
    hello_type               bigint,
    person_name              text
);

create table org (
		id                       uuid      default uuid_generate_v4() not null primary key,
		created_at               timestamp default now(),
		updated_at               timestamp default now(),
		org_name                 text
);

create table users
(
		id                       uuid      default uuid_generate_v4() not null primary key,
		created_at               timestamp default now(),
		updated_at               timestamp default now(),
		username                text,
		first_name               text,
		last_name                text,
		city                     text,
		zip                      text,
		email                    text,
		job_title                text,
		org_id                   uuid references org(id), 
);
