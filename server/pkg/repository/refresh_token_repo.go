package repository

import (
	"context"
	"fmt"
	"time"

	"dployr.io/pkg/models"
	"github.com/jmoiron/sqlx"
)


type RefreshTokenRepo struct {
	*Repository[models.RefreshToken]
}

func NewRefreshTokenRepo(db *sqlx.DB) *RefreshTokenRepo {
	return &RefreshTokenRepo{
		Repository: NewRepository[models.RefreshToken](db, "refresh_tokens"),
	}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, p models.RefreshToken) error {
	return r.Repository.Create(ctx, p)
}

func (r *RefreshTokenRepo) GetByToken(ctx context.Context, token string) (models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.GetContext(ctx, &token, "SELECT * FROM refresh_tokens WHERE token = ?", token)
	if err != nil {
		return rt, err
	}
	return rt, nil
}

func (r *RefreshTokenRepo) ConsumeToken(ctx context.Context, token string) (models.RefreshToken, error) {
	rt, err := r.GetByToken(ctx, token) 
	
	if err != nil {
		return rt, err
	}
	
	if rt.Used {
		return rt, fmt.Errorf("code already used")
	}
	
	if time.Now().After(rt.ExpiresAt) {
		return rt, fmt.Errorf("code expired")
	}
	
	_, err = r.db.
		ExecContext(ctx, "UPDATE refresh_tokens SET used = true WHERE token = ?", token)

	if err != nil {
		return rt, err
	}

	return rt, nil
}

