-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS logs (
  id           TEXT      PRIMARY KEY,
  project_id   TEXT      NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  host         TEXT      NOT NULL,
  message      TEXT      NOT NULL,
  created_at   DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
  type         TEXT      NOT NULL,
  level        TEXT,
  status       TEXT
);

-- +goose Down
DROP TABLE IF EXISTS logs;
