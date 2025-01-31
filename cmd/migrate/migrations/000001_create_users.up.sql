CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    email citext NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    password bytea NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT now()
);