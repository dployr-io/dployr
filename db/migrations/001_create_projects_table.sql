-- +goose Up
CREATE TABLE IF NOT EXISTS projects (
    id            TEXT      PRIMARY KEY,
    name          TEXT      NOT NULL,
    git_repo      TEXT      NOT NULL,
    domain        TEXT      NOT NULL,
    provider      TEXT      NOT NULL,
    environment   TEXT,     -- JSONB → TEXT
    created_at    DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_deployed DATETIME,
    deployment_url TEXT,
    status        TEXT      NOT NULL,
    host_configs  TEXT      -- JSONB → TEXT
);

-- +goose Down
DROP TABLE IF EXISTS projects;
