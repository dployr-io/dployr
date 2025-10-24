package store

import (
	"context"
	"database/sql"
	"dployr/pkg/store"
	"time"

	"github.com/oklog/ulid/v2"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (u *UserStore) FindOrCreateUser(email string) (*store.User, error) {
	user := &store.User{}
	stmt, err := u.db.Prepare("SELECT id, email FROM users WHERE email = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(email).Scan(&user.ID, &user.Email)
	if err == sql.ErrNoRows {
		id := ulid.Make().String()
		now := time.Now().UTC()
		insertStmt, err := u.db.Prepare("INSERT INTO users (id, email, created_at, updated_at) VALUES (?, ?, ?, ?)")
		if err != nil {
			return nil, err
		}
		defer insertStmt.Close()
		_, err = insertStmt.Exec(id, email, now, now)
		if err != nil {
			return nil, err
		}
		user.ID = id
		user.Email = email
		user.CreatedAt = now
		user.UpdatedAt = now
		return user, nil
	}
	return user, err
}

func (u *UserStore) SaveMagicToken(email, token string, expiry time.Time) error {
	_, err := u.db.Exec("UPDATE users SET magic_token=?, magic_token_expiry=? WHERE email=?", token, expiry, email)
	return err
}

func (u *UserStore) ValidateMagicToken(token string) (*store.User, error) {
	stmt, err := u.db.Prepare(`
		SELECT id, email FROM users 
		WHERE magic_token = ? AND magic_token_expiry > datetime('now')
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	user := &store.User{}
	err = stmt.QueryRow(token).Scan(&user.ID, &user.Email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u UserStore) GetUserByEmail(ctx context.Context, email string) (*store.User, error) {
	stmt, err := u.db.PrepareContext(ctx, `
		SELECT id, name, email, password, magic_token, magic_token_expiry, created_at, updated_at
		FROM users WHERE email = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, email)

	var user store.User
	err = row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.MagicToken, &user.MagicTokenExpiry, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
