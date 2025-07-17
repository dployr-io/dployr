-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS projects (
    id             TEXT      PRIMARY KEY,
    name           TEXT      NOT NULL,
    description    TEXT,
    logo           TEXT,
    git_repo       TEXT      NOT NULL,
    domains        JSON      NOT NULL DEFAULT '[]',
    environment    JSON      NOT NULL DEFAULT '[]',
    created_at     DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_deployed  DATETIME,
    deployment_url TEXT,
    status         TEXT      NOT NULL,
    host_configs   JSON      NOT NULL DEFAULT '[]'
);

-- +goose Down
DROP TABLE IF EXISTS projects;