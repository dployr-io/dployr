-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS deployments (
  commit_hash  TEXT      PRIMARY KEY,
  project_id   TEXT      NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  branch       TEXT      NOT NULL,
  duration     INTEGER   NOT NULL,
  message      TEXT      NOT NULL,
  created_at   DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
  status       TEXT
);

-- +goose Down
DROP TABLE IF EXISTS deployments;
