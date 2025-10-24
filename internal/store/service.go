package store

import (
	"context"
	"database/sql"

	"dployr/pkg/store"
)

type ServiceStore struct {
	db *sql.DB
}

func (s ServiceStore) CreateService(ctx context.Context, svc *store.Service) error {
	stmt, err := s.db.PrepareContext(ctx, `
		INSERT INTO services
		(id, name, source, runtime, runtime_version, run_cmd, build_cmd, port, working_dir,
		static_dir, image, env_vars, secrets, status, project_id, remote_id,
		created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, svc.ID, svc.Name, svc.Source, svc.Runtime, svc.RuntimeVersion, svc.RunCmd, svc.BuildCmd, svc.Port,
		svc.WorkingDir, svc.StaticDir, svc.Image, svc.EnvVars, svc.Secrets, svc.Status,
		svc.ProjectID, svc.RemoteID, svc.CreatedAt, svc.UpdatedAt)
	return err
}

func (s ServiceStore) GetService(ctx context.Context, id string) (*store.Service, error) {
	stmt, err := s.db.PrepareContext(ctx, `
		SELECT id, name, source, runtime, runtime_version, run_cmd, port, working_dir,
		       static_dir, image, env_vars, secrets, status, project_id, remote_id,
		       created_at, updated_at
		FROM services WHERE id = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)

	var svc store.Service
	err = row.Scan(
		&svc.ID, &svc.Name, &svc.Source, &svc.Runtime, &svc.RuntimeVersion, &svc.RunCmd,
		&svc.Port, &svc.WorkingDir, &svc.StaticDir, &svc.Image, &svc.EnvVars, &svc.Secrets,
		&svc.Status, &svc.ProjectID, &svc.RemoteID, &svc.CreatedAt, &svc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &svc, nil
}
