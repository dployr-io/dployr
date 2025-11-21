package store

import (
	"context"
	"database/sql"
	"time"

	"dployr/pkg/store"
)

type ServiceStore struct {
	db *sql.DB
	ds *DeploymentStore
}

func NewServiceStore(db *sql.DB, ds *DeploymentStore) *ServiceStore {
	return &ServiceStore{
		db: db,
		ds: ds,
	}
}

func (s ServiceStore) CreateService(ctx context.Context, svc *store.Service) (*store.Service, error) {
	stmt, err := s.db.PrepareContext(ctx, `
		INSERT INTO services
		(id, name, source, runtime, runtime_version, run_cmd, build_cmd, working_dir,
		static_dir, image, remote_url, remote_branch, remote_commit_hash, deployment_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	createdAt := svc.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	updatedAt := svc.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}

	_, err = stmt.ExecContext(ctx, svc.ID, svc.Name, svc.Source, svc.Runtime, svc.RuntimeVersion, svc.RunCmd, svc.BuildCmd,
		svc.WorkingDir, svc.StaticDir, svc.Image, svc.Remote, svc.Branch, svc.CommitHash, svc.DeploymentId, createdAt.Unix(), updatedAt.Unix())
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func (s ServiceStore) GetService(ctx context.Context, id string) (*store.Service, error) {
	stmt, err := s.db.PrepareContext(ctx, `
		SELECT id, name, source, runtime, runtime_version, run_cmd, working_dir,
		       static_dir, image, remote_url, remote_branch, remote_commit_hash, deployment_id, created_at, updated_at
		FROM services WHERE id = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)

	var svc store.Service
	var createdAtUnix, updatedAtUnix int64
	err = row.Scan(
		&svc.ID, &svc.Name, &svc.Source, &svc.Runtime, &svc.RuntimeVersion, &svc.RunCmd,
		&svc.WorkingDir, &svc.StaticDir, &svc.Image, &svc.Remote, &svc.Branch,
		&svc.CommitHash, &svc.DeploymentId, &createdAtUnix, &updatedAtUnix,
	)
	if err != nil {
		return nil, err
	}
	svc.CreatedAt = time.Unix(createdAtUnix, 0)
	svc.UpdatedAt = time.Unix(updatedAtUnix, 0)

	if svc.DeploymentId != "" {
		deployment, err := s.ds.GetDeployment(ctx, svc.DeploymentId)
		if err == nil {
			svc.Blueprint = &deployment.Blueprint
		}
	}

	return &svc, nil
}

func (s ServiceStore) ListServices(ctx context.Context, limit, offset int) ([]*store.Service, error) {
	stmt, err := s.db.PrepareContext(ctx, `
		SELECT id, name, source, runtime, runtime_version, run_cmd, port, working_dir,
		       static_dir, image, remote_url, remote_branch, remote_commit_hash, deployment_id, created_at, updated_at
		FROM services
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

	var services []*store.Service
	for rows.Next() {
		var svc store.Service
		var createdAtUnix, updatedAtUnix int64
		err := rows.Scan(
			&svc.ID, &svc.Name, &svc.Source, &svc.Runtime, &svc.RuntimeVersion, &svc.RunCmd,
			&svc.WorkingDir, &svc.StaticDir, &svc.Image, &svc.Remote, &svc.Branch,
			&svc.CommitHash, &svc.DeploymentId, &createdAtUnix, &updatedAtUnix,
		)
		if err != nil {
			return nil, err
		}
		svc.CreatedAt = time.Unix(createdAtUnix, 0)
		svc.UpdatedAt = time.Unix(updatedAtUnix, 0)

		if svc.DeploymentId != "" {
			deployment, err := s.ds.GetDeployment(ctx, svc.DeploymentId)
			if err == nil {
				svc.Blueprint = &deployment.Blueprint
			}
		}

		services = append(services, &svc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

func (s ServiceStore) UpdateService(ctx context.Context, svc *store.Service) error {
	stmt, err := s.db.PrepareContext(ctx, `
		UPDATE services 
		SET name = ?, source = ?, runtime = ?, runtime_version = ?, run_cmd = ?, build_cmd = ?, 
		    working_dir = ?, static_dir = ?, image = ?, remote_url = ?, remote_branch = ?, 
			remote_commit_hash = ?, deployment_id = ?, updated_at = ? 
		WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, svc.Name, svc.Source, svc.Runtime, svc.RuntimeVersion, svc.RunCmd, svc.BuildCmd,
		svc.WorkingDir, svc.StaticDir, svc.Image, svc.Remote, svc.Branch, svc.CommitHash, svc.DeploymentId,
		svc.UpdatedAt, svc.ID)
	return err
}
