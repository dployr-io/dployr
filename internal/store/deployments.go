package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/dployr-io/dployr/pkg/store"
)

type DeploymentStore struct {
	db *sql.DB
}

func NewDeploymentStore(db *sql.DB) *DeploymentStore {
	return &DeploymentStore{db: db}
}

func (ds DeploymentStore) CreateDeployment(ctx context.Context, deployment *store.Deployment) error {
	stmt, err := ds.db.PrepareContext(ctx, `
		INSERT INTO deployments (id, user_id, config, status, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	configJSON, err := json.Marshal(deployment.Blueprint)
	if err != nil {
		return err
	}

	createdAt := deployment.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	updatedAt := deployment.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}

	_, err = stmt.ExecContext(ctx, deployment.ID, deployment.UserId, configJSON, deployment.Status, deployment.Metadata, createdAt.Unix(), updatedAt.Unix())
	return err
}

func (ds DeploymentStore) GetDeployment(ctx context.Context, id string) (*store.Deployment, error) {
	stmt, err := ds.db.PrepareContext(ctx, `
		SELECT id, user_id, config, status, metadata, created_at, updated_at
		FROM deployments WHERE id = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)

	var d store.Deployment
	var configJSON []byte
	var createdAtUnix, updatedAtUnix int64
	err = row.Scan(&d.ID, &d.UserId, &configJSON, &d.Status, &d.Metadata, &createdAtUnix, &updatedAtUnix)
	if err != nil {
		return nil, err
	}
	d.CreatedAt = time.Unix(createdAtUnix, 0)
	d.UpdatedAt = time.Unix(updatedAtUnix, 0)

	if err := json.Unmarshal(configJSON, &d.Blueprint); err != nil {
		return nil, err
	}

	return &d, nil
}

func (ds DeploymentStore) ListDeployments(ctx context.Context, limit, offset int) ([]*store.Deployment, error) {
	stmt, err := ds.db.PrepareContext(ctx, `
		SELECT id, user_id, config, status, metadata, created_at, updated_at
		FROM deployments
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*store.Deployment
	for rows.Next() {
		var d store.Deployment
		var blueprint []byte
		var createdAtUnix, updatedAtUnix int64
		err := rows.Scan(&d.ID, &d.UserId, &blueprint, &d.Status, &d.Metadata, &createdAtUnix, &updatedAtUnix)
		if err != nil {
			return nil, err
		}
		d.CreatedAt = time.Unix(createdAtUnix, 0)
		d.UpdatedAt = time.Unix(updatedAtUnix, 0)
		if err := json.Unmarshal(blueprint, &d.Blueprint); err != nil {
			return nil, err
		}
		deployments = append(deployments, &d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return deployments, nil
}

func (ds DeploymentStore) UpdateDeploymentStatus(ctx context.Context, id, status string) error {
	stmt, err := ds.db.PrepareContext(ctx, `
		UPDATE deployments SET status = ?, updated_at = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, status, time.Now().Unix(), id)
	return err
}
