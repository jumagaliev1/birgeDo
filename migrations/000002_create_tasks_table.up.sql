CREATE TABLE IF NOT EXISTS rooms (
    id bigserial PRIMARY KEY,
    title text NOT NULL
);
CREATE TABLE IF NOT EXISTS tasks (
    id bigserial PRIMARY KEY,
    title text NOT NULL,
    room_id bigint NOT NULL REFERENCES rooms ON DELETE CASCADE
);