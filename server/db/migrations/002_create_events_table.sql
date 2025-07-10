-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS events (
    id           TEXT      PRIMARY KEY,
    type         TEXT      NOT NULL,
    aggregate_id TEXT      NOT NULL,
    user_id      TEXT      NOT NULL,
    timestamp    DATETIME  NOT NULL,               -- TIMESTAMPTZ → DATETIME
    data         JSON,                             -- JSONB → JSON (stored as TEXT)
    metadata     JSON,
    version      INTEGER   NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS events;
