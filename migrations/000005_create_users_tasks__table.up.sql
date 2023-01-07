CREATE TABLE IF NOT EXISTS rooms_users (
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    task_id bigint NOT NULL REFERENCES tasks ON DELETE CASCADE,
    done boolean NOT NULL,
    PRIMARY KEY (user_id, task_id)
);