package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"dployr.io/pkg/models"
)

type Project struct {
	*Repository[models.Project]
}

func NewProjectRepo(db *sqlx.DB) *Project {
	return &Project{
		Repository: NewRepository[models.Project](db, "projects"),
	}
}

func (r *Project) Create(ctx context.Context, p *models.Project) (*models.Project, error) {
	return p, r.Repository.Create(ctx, p)
}

func (r *Project) Update(ctx context.Context, p *models.Project) error {
	return r.Repository.Update(ctx, p)
}

func (r *Project) Upsert(ctx context.Context, p *models.Project, conflictCols []string, updateCols []string) error {
	return r.Repository.Upsert(ctx, p, conflictCols, updateCols)
}

func (r *Project) Delete(ctx context.Context, id any) error {
	return r.Repository.Delete(ctx, id)
}

func (r *Project) GetByID(ctx context.Context, id any) (*models.Project, error) {
	return r.Repository.GetByID(ctx, id)
}
