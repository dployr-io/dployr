-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users (
    id             TEXT      PRIMARY KEY,
    name           TEXT      NOT NULL,
    email          TEXT      NOT NULL UNIQUE,
    role           TEXT      NOT NULL,
    avatar         TEXT,
    created_at     DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_users_email ON users(email);

-- +goose Down
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;