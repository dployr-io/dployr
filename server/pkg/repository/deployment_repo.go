package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"dployr.io/pkg/models"
)

type DeploymentRepo struct {
	*Repository[models.Deployment]
}

func NewDeploymentRepo(db *sqlx.DB) *DeploymentRepo {
	return &DeploymentRepo{
		Repository: NewRepository[models.Deployment](db, "deployments"),
	}
}

func (r *DeploymentRepo) Create(ctx context.Context, p models.Deployment) error {
	if p.CommitHash == "" {
        return fmt.Errorf("commit hash is required")
    }

	if p.Branch == "" {
        return fmt.Errorf("branch is required")
    }

    const insertSQL = `
    INSERT INTO deployments (commit_hash, message, branch, status)
    VALUES (:commit_hash, :message, :branch, "pending")
    `
    _, err := r.db.NamedExecContext(ctx, insertSQL, p)
    if err != nil {
        return fmt.Errorf("create deployment failed: %w", err)
    }

	return nil
}

func (r *DeploymentRepo) Update(ctx context.Context, p models.Deployment) error {
	return r.Repository.Update(ctx, p)
}

func (r *DeploymentRepo) Upsert(ctx context.Context, p models.Deployment, conflictCols []string, updateCols []string) error {
	return r.Repository.Upsert(ctx, p, conflictCols, updateCols)
}

func (r *DeploymentRepo) Delete(ctx context.Context, id string) error {
	return r.Repository.Delete(ctx, id)
}

func (r *DeploymentRepo) GetByID(ctx context.Context, id string) (models.Deployment, error) {
	return r.Repository.GetByID(ctx, id)
}

func (r *DeploymentRepo) GetAll(ctx context.Context) ([]models.Deployment, error) {
	d := []models.Deployment{}  
	q := `SELECT * FROM deployments ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &d, q)
	return d, err
}

func (r *DeploymentRepo) GetByCommitHash(ctx context.Context, commitHash string) (models.Deployment, error) {
	var d models.Deployment
	q := `SELECT * FROM deployments WHERE commit_hash = ? LIMIT 1`
    err := r.db.GetContext(ctx, &d, q, commitHash)
    if err != nil {
        return models.Deployment{}, err
    }
    return d, nil
}

func (r *DeploymentRepo) GetByProjectID(ctx context.Context, projectId string) ([]models.Deployment, error) {
	d := []models.Deployment{}  
	q := `SELECT * FROM deployments WHERE project_id = ? ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &d, q, projectId)
	return d, err
}

