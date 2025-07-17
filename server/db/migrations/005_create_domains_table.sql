-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS domains (
  id                  TEXT      PRIMARY KEY,
  project_id          TEXT      NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  subdomain           TEXT      NOT NULL,
  provider            TEXT      NOT NULL,
  auto_setup_available BOOLEAN  NOT NULL,
  manual_records      TEXT,
  verified            BOOLEAN  NOT NULL,
  updated_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS domains;