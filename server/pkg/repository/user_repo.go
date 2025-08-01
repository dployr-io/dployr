package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"dployr.io/pkg/models"
)

type UserRepo struct {
	*Repository[models.User]
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{
		Repository: NewRepository[models.User](db, "users"),
	}
}

func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
	if u.Id == "" {
		u.Id = uuid.NewString()
	}

	const insertSQL = `
    INSERT INTO users (id, name, email, role)
    VALUES (:id, :name, :email, :role)
    `
	_, err := r.db.NamedExecContext(ctx, insertSQL, u)
	if err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}

	return nil
}

func (r *UserRepo) Update(ctx context.Context, p models.User) error {
	return r.Repository.Update(ctx, p)
}

func (r *UserRepo) Upsert(ctx context.Context, p models.User, conflictCols []string, updateCols []string) error {
	return r.Repository.Upsert(ctx, p, conflictCols, updateCols)
}

func (r *UserRepo) Delete(ctx context.Context, id any) error {
	return r.Repository.Delete(ctx, id)
}

func (r *UserRepo) GetByID(ctx context.Context, id any) (models.User, error) {
	return r.Repository.GetByID(ctx, id)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE email = ?", email)
	if err != nil {
		return user, err
	}
	return user, nil
}
