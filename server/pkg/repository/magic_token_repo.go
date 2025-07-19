package repository

import (
	"context"
	
	"github.com/jmoiron/sqlx"
	
	"dployr.io/pkg/models"
)

type MagicTokenRepo struct {
	*Repository[models.MagicToken]
}

func NewMagicTokenRepo(db *sqlx.DB) *MagicTokenRepo {
	return &MagicTokenRepo{
		Repository: NewRepository[models.MagicToken](db, "magic_tokens"),
	}
}

func (r *MagicTokenRepo) Create(ctx context.Context, p models.MagicToken) error {
	return r.Repository.Create(ctx, p)
}

func (r *MagicTokenRepo) Update(ctx context.Context, p models.MagicToken) error {
	return r.Repository.Update(ctx, p)
}

func (r *MagicTokenRepo) Upsert(ctx context.Context, p models.MagicToken, conflictCols []string, updateCols []string) error {
	return r.Repository.Upsert(ctx, p, conflictCols, updateCols)
}

func (r *MagicTokenRepo) Delete(ctx context.Context, id any) error {
	return r.Repository.Delete(ctx, id)
}

func (r *MagicTokenRepo) GetByID(ctx context.Context, id any) (models.MagicToken, error) {
	return r.Repository.GetByID(ctx, id)
}

func (r *MagicTokenRepo) GetByEmail(ctx context.Context, email string) (models.MagicToken, error) {
	var token models.MagicToken
	err := r.db.GetContext(ctx, &token, "SELECT * FROM magic_tokens WHERE email = ?", email)
	if err != nil {
		return token, err
	}
	return token, nil
}

func (r *MagicTokenRepo) ConsumeCode(ctx context.Context, email, code string) (models.MagicToken, error) {
	var mt models.MagicToken
	_, err := r.db.
		ExecContext(ctx, "UPDATE magic_tokens SET used = true WHERE email = ? AND code = ?", email, code)
		
	if err != nil {
		return mt, err
	}

	magicToken, err := r.GetByEmail(ctx, email) 

	if err != nil {
		return magicToken, err
	}

	return magicToken, nil
}

