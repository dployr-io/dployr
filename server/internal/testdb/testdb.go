// server/internal/testdb/testdb.go
package testdb

import (
	"context"
	"database/sql"
	"io/fs"
	"testing"

	"dployr.io/pkg/models"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	sqlite3 "modernc.org/sqlite"
)

func init() {
	// Register sqlite3 driver only if not already registered
	defer func() {
		if r := recover(); r != nil {
			// Driver already registered, ignore
		}
	}()
	sql.Register("sqlite3", &sqlite3.Driver{})
}

// SetupTestDB creates a test database with migrations
func SetupTestDB(t *testing.T, migrations fs.ReadFileFS) *sqlx.DB {
	t.Helper()

	db, err := sqlx.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Run migrations
	runTestMigrations(t, db, migrations)

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func runTestMigrations(t *testing.T, db *sqlx.DB, migrations fs.ReadFileFS) {
	t.Helper()

	// Simple migration runner for tests
	migrationFiles := []string{
		"001_init.sql",
		"002_users.sql",
		"003_projects.sql",
		"004_deployments.sql",
		"005_logs.sql",
	}

	for _, file := range migrationFiles {
		content, err := migrations.ReadFile("db/migrations/" + file)
		if err != nil {
			continue // Skip missing files
		}

		_, err = db.Exec(string(content))
		require.NoError(t, err, "failed to run migration %s", file)
	}
}

// UserRepo interface for testing - minimal interface to avoid import cycles
type UserRepo interface {
	Create(ctx context.Context, user *models.User) error
}

// CreateTestUser creates a test user
func CreateTestUser(t *testing.T, repo UserRepo) *models.User {
	t.Helper()

	user := &models.User{
		Name:  "Test User",
		Email: "test@example.com",
		Role:  "user",
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	return user
}
