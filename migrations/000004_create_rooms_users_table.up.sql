CREATE TABLE IF NOT EXISTS rooms_users (
    room_id bigint NOT NULL REFERENCES rooms ON DELETE CASCADE,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    PRIMARY KEY (room_id, user_id)
);