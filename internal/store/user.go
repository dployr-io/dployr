package store

import (
	"context"
	"database/sql"
	"time"

	"dployr/pkg/store"

	"github.com/oklog/ulid/v2"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (u *UserStore) FindOrCreateUser(email string, role store.Role) (*store.User, error) {
	tx, err := u.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user, err := u.findUser(tx, email)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err != sql.ErrNoRows {
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return user, nil
	}

	role, err = u.updateRole(tx, role)
	if err != nil {
		return nil, err
	}

	user, err = u.createUser(tx, email, role)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserStore) findUser(tx *sql.Tx, email string) (*store.User, error) {
	user := &store.User{}
	stmt, err := tx.Prepare("SELECT id, email, role, created_at, updated_at FROM users WHERE email = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(email).Scan(&user.ID, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func (u *UserStore) updateRole(tx *sql.Tx, role store.Role) (store.Role, error) {
	if role != store.RoleOwner {
		return role, nil
	}

	var ownerCount int
	err := tx.QueryRow("SELECT COUNT(*) FROM users WHERE role = ?", store.RoleOwner).Scan(&ownerCount)
	if err != nil {
		return role, err
	}

	if ownerCount > 0 {
		return store.RoleViewer, nil
	}

	return role, nil
}

func (u *UserStore) createUser(tx *sql.Tx, email string, role store.Role) (*store.User, error) {
	id := ulid.Make().String()
	now := time.Now().UTC()

	insertStmt, err := tx.Prepare("INSERT INTO users (id, email, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}
	defer insertStmt.Close()

	_, err = insertStmt.Exec(id, email, role, now, now)
	if err != nil {
		return nil, err
	}

	return &store.User{
		ID:        id,
		Email:     email,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (u *UserStore) UpdateUserRole(ctx context.Context, email string, role store.Role) error {
	stmt, err := u.db.PrepareContext(ctx, "UPDATE users SET role = ?, updated_at = ? WHERE email = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, role, time.Now().UTC(), email)
	return err
}

func (u UserStore) GetUserByEmail(ctx context.Context, email string) (*store.User, error) {
	stmt, err := u.db.PrepareContext(ctx, `
		SELECT id, email, role, created_at, updated_at
		FROM users WHERE email = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, email)

	var user store.User
	err = row.Scan(&user.ID, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserStore) IsOwner() (bool, error) {
	// Check if owner exists
	var count int
	err := u.db.QueryRow("SELECT COUNT(*) FROM users WHERE role = ?", store.RoleOwner).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil // True if owner exists
}
