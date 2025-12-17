// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/dployr-io/dployr/pkg/store"
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

func (s ServiceStore) createService(ctx context.Context, svc *store.Service) (*store.Service, error) {
	stmt, err := s.db.PrepareContext(ctx, `
		INSERT INTO services
		(id, name, description, source, runtime, runtime_version, run_cmd, build_cmd, working_dir,
		static_dir, image, remote_url, remote_branch, remote_commit_hash, deployment_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
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

	_, err = stmt.ExecContext(ctx, svc.ID, svc.Name, svc.Description, svc.Source, svc.Runtime, svc.RuntimeVersion, svc.RunCmd, svc.BuildCmd,
		svc.WorkingDir, svc.StaticDir, svc.Image, svc.Remote, svc.Branch, svc.CommitHash, svc.DeploymentId, createdAt.Unix(), updatedAt.Unix())
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func (s ServiceStore) GetService(ctx context.Context, id string) (*store.Service, error) {
	stmt, err := s.db.PrepareContext(ctx, `
		SELECT id, name, description, source, runtime, runtime_version, run_cmd, build_cmd, working_dir,
		       static_dir, image, remote_url, remote_branch, remote_commit_hash, deployment_id, created_at, updated_at
		FROM services WHERE id = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)

	var svc store.Service
	var createdAtUnix, updatedAtUnix int64
	var description, runCmd, buildCmd, staticDir, image, remoteURL, remoteBranch, remoteCommitHash, deploymentID sql.NullString
	err = row.Scan(
		&svc.ID, &svc.Name, &description, &svc.Source, &svc.Runtime, &svc.RuntimeVersion, &runCmd, &buildCmd,
		&svc.WorkingDir, &staticDir, &image, &remoteURL, &remoteBranch,
		&remoteCommitHash, &deploymentID, &createdAtUnix, &updatedAtUnix,
	)
	if err != nil {
		return nil, err
	}
	svc.Description = description.String
	svc.RunCmd = runCmd.String
	svc.BuildCmd = buildCmd.String
	svc.StaticDir = staticDir.String
	svc.Image = image.String
	svc.Remote = remoteURL.String
	svc.Branch = remoteBranch.String
	svc.CommitHash = remoteCommitHash.String
	svc.DeploymentId = deploymentID.String
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
		SELECT id, name, description, source, runtime, runtime_version, run_cmd, build_cmd, working_dir,
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
		var description, runCmd, buildCmd, staticDir, image, remoteURL, remoteBranch, remoteCommitHash, deploymentID sql.NullString
		err := rows.Scan(
			&svc.ID, &svc.Name, &description, &svc.Source, &svc.Runtime, &svc.RuntimeVersion, &runCmd, &buildCmd,
			&svc.WorkingDir, &staticDir, &image, &remoteURL, &remoteBranch,
			&remoteCommitHash, &deploymentID, &createdAtUnix, &updatedAtUnix,
		)
		if err != nil {
			return nil, err
		}
		svc.Description = description.String
		svc.RunCmd = runCmd.String
		svc.BuildCmd = buildCmd.String
		svc.StaticDir = staticDir.String
		svc.Image = image.String
		svc.Remote = remoteURL.String
		svc.Branch = remoteBranch.String
		svc.CommitHash = remoteCommitHash.String
		svc.DeploymentId = deploymentID.String
		svc.CreatedAt = time.Unix(createdAtUnix, 0)
		svc.UpdatedAt = time.Unix(updatedAtUnix, 0)

		services = append(services, &svc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Fetch blueprints after closing rows to avoid SQLite deadlock (MaxOpenConns=1)
	for _, svc := range services {
		if svc.DeploymentId != "" {
			deployment, err := s.ds.GetDeployment(ctx, svc.DeploymentId)
			if err == nil {
				svc.Blueprint = &deployment.Blueprint
			}
		}
	}

	return services, nil
}

func (s ServiceStore) updateService(ctx context.Context, svc *store.Service) error {
	stmt, err := s.db.PrepareContext(ctx, `
		UPDATE services 
		SET name = ?, description = ?, source = ?, runtime = ?, runtime_version = ?, run_cmd = ?, build_cmd = ?, 
		    working_dir = ?, static_dir = ?, image = ?, remote_url = ?, remote_branch = ?, 
			remote_commit_hash = ?, deployment_id = ?, updated_at = ? 
		WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, svc.Name, svc.Description, svc.Source, svc.Runtime, svc.RuntimeVersion, svc.RunCmd, svc.BuildCmd,
		svc.WorkingDir, svc.StaticDir, svc.Image, svc.Remote, svc.Branch, svc.CommitHash, svc.DeploymentId,
		svc.UpdatedAt, svc.ID)
	return err
}

// SaveService creates a new service or updates an existing one.
func (s ServiceStore) SaveService(ctx context.Context, svc *store.Service) (*store.Service, error) {
	existing, err := s.GetService(ctx, svc.ID)
	if err == nil && existing != nil {
		svc.CreatedAt = existing.CreatedAt
		svc.UpdatedAt = time.Now()
		if err := s.updateService(ctx, svc); err != nil {
			return nil, err
		}
		return svc, nil
	}

	return s.createService(ctx, svc)
}

// DeleteService removes a service by ID.
func (s ServiceStore) DeleteService(ctx context.Context, id string) error {
	stmt, err := s.db.PrepareContext(ctx, `DELETE FROM services WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	return err
}
