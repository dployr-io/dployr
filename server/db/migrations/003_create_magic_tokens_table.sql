-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS magic_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    code TEXT NOT NULL UNIQUE,
    name TEXT,
    expires_at DATETIME NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_magic_tokens_token ON magic_tokens(code);
CREATE UNIQUE INDEX idx_magic_tokens_email ON magic_tokens(email);

-- +goose Down
DROP INDEX IF EXISTS idx_magic_tokens_email;
DROP INDEX IF EXISTS idx_magic_tokens_token;
DROP TABLE IF EXISTS magic_tokens;