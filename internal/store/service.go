package store

import (
	"context"
	"database/sql"

	"dployr/pkg/store"
)

type ServiceStore struct {
	db *sql.DB
}

func NewServiceStore(db *sql.DB) *ServiceStore {
	return &ServiceStore{db: db}
}

func (s ServiceStore) CreateService(ctx context.Context, svc *store.Service) (*store.Service, error) {
	stmt, err := s.db.PrepareContext(ctx, `
		INSERT INTO services
		(id, name, source, runtime, runtime_version, run_cmd, build_cmd, port, working_dir,
		static_dir, image, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, svc.ID, svc.Name, svc.Source, svc.Runtime, svc.RuntimeVersion, svc.RunCmd, svc.BuildCmd, svc.Port,
		svc.WorkingDir, svc.StaticDir, svc.Image, svc.CreatedAt, svc.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func (s ServiceStore) GetService(ctx context.Context, id string) (*store.Service, error) {
	stmt, err := s.db.PrepareContext(ctx, `
		SELECT id, name, source, runtime, runtime_version, run_cmd, port, working_dir,
		       static_dir, image, created_at, updated_at
		FROM services WHERE id = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)

	var svc store.Service
	err = row.Scan(
		&svc.ID, &svc.Name, &svc.Source, &svc.Runtime, &svc.RuntimeVersion, &svc.RunCmd,
		&svc.Port, &svc.WorkingDir, &svc.StaticDir, &svc.Image, &svc.CreatedAt, &svc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

func (s ServiceStore) ListServices(ctx context.Context, limit, offset int) ([]*store.Service, error) {
	stmt, err := s.db.PrepareContext(ctx, `
		SELECT id, name, source, runtime, runtime_version, run_cmd, port, working_dir,
		       static_dir, image, created_at, updated_at
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
		err := rows.Scan(
			&svc.ID, &svc.Name, &svc.Source, &svc.Runtime, &svc.RuntimeVersion, &svc.RunCmd,
			&svc.Port, &svc.WorkingDir, &svc.StaticDir, &svc.Image, &svc.CreatedAt, &svc.UpdatedAt,
		)
		if err != nil {
			return nil, err
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
		    port = ?, working_dir = ?, static_dir = ?, image = ?, updated_at = ? 
		WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, svc.Name, svc.Source, svc.Runtime, svc.RuntimeVersion, svc.RunCmd, svc.BuildCmd,
		svc.Port, svc.WorkingDir, svc.StaticDir, svc.Image, svc.UpdatedAt, svc.ID)
	return err
}
