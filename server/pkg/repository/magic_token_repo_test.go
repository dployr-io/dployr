// server/pkg/repository/magic_token_repo_test.go
package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dployr.io/internal/testdb"
	"dployr.io/pkg/models"
)

func TestMagicTokenRepo_Create(t *testing.T) {
	db := testdb.SetupTestDB(t, testdb.MockMigrations{})
	repo := NewMagicTokenRepo(db)

	token := &models.MagicToken{
		Code:      "ABC123",
		Email:     "test@example.com",
		Name:      "Test User",
		Used:      false,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	err := repo.Create(context.Background(), token)
	require.NoError(t, err)
	assert.NotEmpty(t, token.Id)
}

func TestMagicTokenRepo_GetByEmail(t *testing.T) {
	db := testdb.SetupTestDB(t, testdb.MockMigrations{})
	repo := NewMagicTokenRepo(db)

	// Create test token
	token := &models.MagicToken{
		Code:      "ABC123",
		Email:     "test@example.com",
		Name:      "Test User",
		Used:      false,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	err := repo.Create(context.Background(), token)
	require.NoError(t, err)

	// Test retrieval
	retrieved, err := repo.GetByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, token.Code, retrieved.Code)
	assert.Equal(t, token.Email, retrieved.Email)
	assert.Equal(t, token.Name, retrieved.Name)
}

func TestMagicTokenRepo_ConsumeCode(t *testing.T) {
	db := testdb.SetupTestDB(t, testdb.MockMigrations{})
	repo := NewMagicTokenRepo(db)

	// Create test token
	token := &models.MagicToken{
		Code:      "ABC123",
		Email:     "test@example.com",
		Name:      "Test User",
		Used:      false,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	err := repo.Create(context.Background(), token)
	require.NoError(t, err)

	tests := []struct {
		name    string
		email   string
		code    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid code",
			email:   "test@example.com",
			code:    "ABC123",
			wantErr: false,
		},
		{
			name:    "already used code",
			email:   "test@example.com",
			code:    "ABC123",
			wantErr: true,
			errMsg:  "code already used",
		},
		{
			name:    "invalid email",
			email:   "invalid@example.com",
			code:    "ABC123",
			wantErr: true,
		},
		{
			name:    "invalid code",
			email:   "test@example.com",
			code:    "INVALID",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := repo.ConsumeCode(context.Background(), tt.email, tt.code)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMagicTokenRepo_ConsumeCode_Expired(t *testing.T) {
	db := testdb.SetupTestDB(t, testdb.MockMigrations{})
	repo := NewMagicTokenRepo(db)

	// Create expired token
	token := &models.MagicToken{
		Code:      "EXPIRED",
		Email:     "expired@example.com",
		Name:      "Expired User",
		Used:      false,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
	}
	err := repo.Create(context.Background(), token)
	require.NoError(t, err)

	// Try to consume expired code
	_, err = repo.ConsumeCode(context.Background(), "expired@example.com", "EXPIRED")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "code expired")
}
