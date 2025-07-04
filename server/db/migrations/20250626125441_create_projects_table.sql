-- +goose Up
CREATE TABLE IF NOT EXISTS projects (
    id            TEXT      PRIMARY KEY,
    name          TEXT      NOT NULL,
    git_repo      TEXT      NOT NULL,
    domain        TEXT      NOT NULL,
    provider      TEXT      NOT NULL,
    environment   JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_deployed TIMESTAMPTZ,
    deployment_url TEXT,
    status        TEXT      NOT NULL,
    host_configs  JSONB
);

-- +goose Down
DROP TABLE events;
