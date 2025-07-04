-- +goose Up
CREATE TABLE IF NOT EXISTS events (
    id            TEXT      PRIMARY KEY,
    type          TEXT      NOT NULL,
    aggregate_id  TEXT      NOT NULL,
    user_id       TEXT      NOT NULL,
    timestamp     TIMESTAMPTZ NOT NULL,
    data          JSONB,
    metadata      JSONB,
    version       INTEGER   NOT NULL
);

-- +goose Down
DROP TABLE events;
