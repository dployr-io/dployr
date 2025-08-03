// server/pkg/repository/user_repo_test.go
package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dployr.io/internal/testdb"
	"dployr.io/pkg/models"
)

func TestUserRepo_Create(t *testing.T) {
	db := testdb.SetupTestDB(t, testdb.MockMigrations{})
	repo := NewUserRepo(db)

	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
	}{
		{
			name: "valid user",
			user: &models.User{
				Name:  "John Doe",
				Email: "john@example.com",
				Role:  "user",
			},
			wantErr: false,
		},
		{
			name: "duplicate email",
			user: &models.User{
				Name:  "Jane Doe",
				Email: "john@example.com", // Same email
				Role:  "admin",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(context.Background(), tt.user)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, tt.user.Id)
				assert.NotZero(t, tt.user.CreatedAt)
			}
		})
	}
}

func TestUserRepo_GetByEmail(t *testing.T) {
	db := testdb.SetupTestDB(t, testdb.MockMigrations{})
	repo := NewUserRepo(db)

	// Create test user
	user := testdb.CreateTestUser(t, repo)

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "existing user",
			email:   user.Email,
			wantErr: false,
		},
		{
			name:    "non-existing user",
			email:   "nonexistent@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByEmail(context.Background(), tt.email)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.email, result.Email)
				assert.Equal(t, user.Name, result.Name)
			}
		})
	}
}

func TestUserRepo_GetByID(t *testing.T) {
	db := testdb.SetupTestDB(t, testdb.MockMigrations{})
	repo := NewUserRepo(db)

	// Create test user
	user := testdb.CreateTestUser(t, repo)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "existing user",
			id:      user.Id,
			wantErr: false,
		},
		{
			name:    "non-existing user",
			id:      "non-existent-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(context.Background(), tt.id)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.id, result.Id)
				assert.Equal(t, user.Email, result.Email)
			}
		})
	}
}

func TestUserRepo_Update(t *testing.T) {
	db := testdb.SetupTestDB(t, testdb.MockMigrations{})
	repo := NewUserRepo(db)

	// Create test user
	user := testdb.CreateTestUser(t, repo)

	// Update user
	user.Name = "Updated Name"
	user.Role = "admin"

	err := repo.Update(context.Background(), *user)
	require.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(context.Background(), user.Id)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "admin", updated.Role)
}

func TestUserRepo_Delete(t *testing.T) {
	db := testdb.SetupTestDB(t, testdb.MockMigrations{})
	repo := NewUserRepo(db)

	// Create test user
	user := testdb.CreateTestUser(t, repo)

	// Delete user
	err := repo.Delete(context.Background(), user.Id)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(context.Background(), user.Id)
	assert.Error(t, err)
}
