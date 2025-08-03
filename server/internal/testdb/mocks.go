// server/internal/testdb/mocks.go
package testdb

import (
	"io/fs"
)

// MockMigrations implements embed.FS for testing
type MockMigrations struct{}

// Ensure MockMigrations implements fs.ReadFileFS
var _ fs.ReadFileFS = MockMigrations{}

func (m MockMigrations) Open(name string) (fs.File, error) {
	return nil, fs.ErrNotExist
}

func (m MockMigrations) ReadDir(name string) ([]fs.DirEntry, error) {
	return nil, nil
}

func (m MockMigrations) ReadFile(name string) ([]byte, error) {
	// Return mock migration content based on filename
	switch name {
	case "db/migrations/001_init.sql":
		return []byte(`
			CREATE TABLE IF NOT EXISTS users (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				email TEXT UNIQUE NOT NULL,
				avatar TEXT,
				role TEXT DEFAULT 'user',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);
		`), nil
	case "db/migrations/002_users.sql":
		return []byte(`
			-- Additional user table modifications
		`), nil
	case "db/migrations/003_projects.sql":
		return []byte(`
			CREATE TABLE IF NOT EXISTS projects (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				logo TEXT,
				description TEXT,
				git_repo TEXT NOT NULL,
				domains TEXT,
				environment TEXT,
				deployment_url TEXT,
				last_deployed DATETIME,
				status TEXT DEFAULT 'setup',
				host_configs TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);
		`), nil
	case "db/migrations/004_deployments.sql":
		return []byte(`
			CREATE TABLE IF NOT EXISTS deployments (
				project_id TEXT,
				commit_hash TEXT PRIMARY KEY,
				branch TEXT NOT NULL,
				duration INTEGER DEFAULT 0,
				message TEXT,
				status TEXT DEFAULT 'pending',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);
		`), nil
	case "db/migrations/005_logs.sql":
		return []byte(`
			CREATE TABLE IF NOT EXISTS logs (
				id TEXT PRIMARY KEY,
				project_id TEXT,
				host TEXT,
				message TEXT NOT NULL,
				status TEXT,
				type TEXT,
				level TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);

			CREATE TABLE IF NOT EXISTS magic_tokens (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				code TEXT NOT NULL UNIQUE,
				email TEXT NOT NULL UNIQUE,
				name TEXT,
				used BOOLEAN DEFAULT FALSE,
				expires_at DATETIME NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);

			CREATE TABLE IF NOT EXISTS refresh_tokens (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				token TEXT UNIQUE NOT NULL,
				user_id TEXT NOT NULL,
				used BOOLEAN DEFAULT FALSE,
				expires_at DATETIME NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);

			CREATE TABLE IF NOT EXISTS events (
				id TEXT PRIMARY KEY,
				type TEXT NOT NULL,
				aggregate_id TEXT NOT NULL,
				user_id TEXT,
				timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
				data TEXT,
				metadata TEXT,
				version INTEGER DEFAULT 1
			);
		`), nil
	default:
		return nil, fs.ErrNotExist
	}
}
