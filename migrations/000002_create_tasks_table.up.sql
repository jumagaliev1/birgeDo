CREATE TABLE IF NOT EXISTS users (
                                     id bigserial PRIMARY KEY,
                                     created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL,
    version integer NOT NULL DEFAULT 1
    );
ALTER TABLE users DROP CONSTRAINT users_uc_email;
ALTER TABLE users ADD CONSTRAINT users_uc_email UNIQUE (email);
CREATE TABLE IF NOT EXISTS rooms (
    id bigserial PRIMARY KEY,
    title text NOT NULL
);
CREATE TABLE IF NOT EXISTS tasks (
    id bigserial PRIMARY KEY,
    title text NOT NULL,
    room_id bigint NOT NULL REFERENCES rooms ON DELETE CASCADE
);