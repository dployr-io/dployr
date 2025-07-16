package repository

import (
	"context"
	
	"github.com/jmoiron/sqlx"
	
	"dployr.io/pkg/models"
)

type ProjectRepo struct {
	*Repository[models.Project]
}

func NewProjectRepo(db *sqlx.DB) *ProjectRepo {
	return &ProjectRepo{
		Repository: NewRepository[models.Project](db, "projects"),
	}
}

func (r *ProjectRepo) Create(ctx context.Context, p *models.Project) error {
	return r.Repository.Create(ctx, p)
}

func (r *ProjectRepo) Update(ctx context.Context, p *models.Project) error {
	return r.Repository.Update(ctx, p)
}

func (r *ProjectRepo) Upsert(ctx context.Context, p *models.Project, conflictCols []string, updateCols []string) error {
	return r.Repository.Upsert(ctx, p, conflictCols, updateCols)
}

func (r *ProjectRepo) Delete(ctx context.Context, id any) error {
	return r.Repository.Delete(ctx, id)
}

func (r *ProjectRepo) GetByID(ctx context.Context, id any) (*models.Project, error) {
	return r.Repository.GetByID(ctx, id)
}

func (r *ProjectRepo) GetAll(ctx context.Context) ([]*models.Project, error) {
	var projects []*models.Project
	query := `SELECT * FROM projects ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &projects, query)
	return projects, err
}

