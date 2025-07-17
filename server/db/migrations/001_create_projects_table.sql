-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS projects (
    id             TEXT      PRIMARY KEY,
    name           TEXT      NOT NULL,
    git_repo       TEXT      NOT NULL,
    domain         TEXT      NOT NULL,
    environment    JSON,
    created_at     DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_deployed  DATETIME,
    deployment_url TEXT,
    status         TEXT      NOT NULL,
    host_configs   JSON
);

-- +goose Down
DROP TABLE IF EXISTS projects;