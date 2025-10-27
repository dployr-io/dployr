package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"dployr/pkg/store"
)

type DeploymentStore struct {
	db *sql.DB
}

func NewDeploymentStore(db *sql.DB) *DeploymentStore {
	return &DeploymentStore{db: db}
}

func (d DeploymentStore) CreateDeployment(ctx context.Context, deployment *store.Deployment) error {
	stmt, err := d.db.PrepareContext(ctx, `
		INSERT INTO deployments (id, user_id, config, status, save_spec, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	configJSON, err := json.Marshal(deployment.Cfg)
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx, deployment.ID, deployment.UserId, configJSON, deployment.Status, deployment.SaveSpec, deployment.Metadata, deployment.CreatedAt, deployment.UpdatedAt)
	return err
}

func (d DeploymentStore) GetDeployment(ctx context.Context, id string) (*store.Deployment, error) {
	stmt, err := d.db.PrepareContext(ctx, `
		SELECT id, user_id, config, status, save_spec, metadata, created_at, updated_at
		FROM deployments WHERE id = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)

	var bp store.Deployment
	var configJSON []byte
	err = row.Scan(&bp.ID, &bp.UserId, &configJSON, &bp.Status, &bp.SaveSpec, &bp.Metadata, &bp.CreatedAt, &bp.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configJSON, &bp.Cfg); err != nil {
		return nil, err
	}

	return &bp, nil
}

func (d DeploymentStore) ListDeployments(ctx context.Context, limit, offset int) ([]*store.Deployment, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT id, user_id, config, status, save_spec, metadata, created_at, updated_at
		FROM deployments
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*store.Deployment
	for rows.Next() {
		var bp store.Deployment
		var configJSON []byte
		err := rows.Scan(&bp.ID, &bp.UserId, &configJSON, &bp.Status, &bp.SaveSpec, &bp.Metadata, &bp.CreatedAt, &bp.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(configJSON, &bp.Cfg); err != nil {
			return nil, err
		}
		deployments = append(deployments, &bp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return deployments, nil
}

func (d DeploymentStore) UpdateDeploymentStatus(ctx context.Context, id, status string) error {
	stmt, err := d.db.PrepareContext(ctx, `
		UPDATE deployments SET status = ?, updated_at = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, status, time.Now(), id)
	return err
}
