package db

import (
	"database/sql"
	"dployr/pkg/core/utils"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Open() (*sql.DB, error) {
	dataDir := utils.GetDataDir()
	dbPath := filepath.Join(dataDir, "data.db")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := applyMigrations(db); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}

func Close(DB *sql.DB) error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

func applyMigrations(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (filename TEXT PRIMARY KEY)`); err != nil {
		return err
	}

	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		var exists string
		_ = db.QueryRow(`SELECT filename FROM schema_migrations WHERE filename = ?`, name).Scan(&exists)
		if exists != "" {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		content, err := migrationFiles.ReadFile("migrations/" + name)
		if err != nil {
			_ = tx.Rollback()
			return err
		}

		// execute atomically
		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("error executing %s: %w", name, err)
		}

		if _, err := tx.Exec(`INSERT INTO schema_migrations (filename) VALUES (?)`, name); err != nil {
			_ = tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit failed for %s: %w", name, err)
		}

		fmt.Println("successfully applied migration:", name)
	}

	return nil
}
