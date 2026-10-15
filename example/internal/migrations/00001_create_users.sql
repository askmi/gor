-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id         bigserial   PRIMARY KEY,
    name       text        NOT NULL,
    email      text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Lookups by email are the application's only non-key access path, and the
-- address identifies the user, so the index enforces uniqueness as well.
CREATE UNIQUE INDEX users_email_key ON users (email);

-- +goose Down
DROP TABLE IF EXISTS users CASCADE;
