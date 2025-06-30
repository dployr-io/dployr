-- +goose Up
CREATE TABLE IF NOT EXISTS events (
    id            TEXT      PRIMARY KEY,
    type          TEXT      NOT NULL,
    aggregate_id  TEXT      NOT NULL,
    user_id       TEXT      NOT NULL,
    timestamp     DATETIME  NOT NULL,
    data          TEXT,      -- JSONB → TEXT
    metadata      TEXT,      -- JSONB → TEXT
    version       INTEGER   NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS events;
