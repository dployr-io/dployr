package store

import (
	"context"
	"database/sql"

	"dployr/pkg/store"
)

type ProjectStore struct {
	db *sql.DB
}

func (p ProjectStore) CreateProject(ctx context.Context, project *store.Project) error {
	stmt, err := p.db.PrepareContext(ctx, `
		INSERT INTO projects (id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, project.ID, project.Name, project.Description, project.CreatedAt, project.UpdatedAt)
	return err
}

func (p ProjectStore) GetProject(ctx context.Context, id string) (*store.Project, error) {
	stmt, err := p.db.PrepareContext(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM projects WHERE id = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)

	var project store.Project
	err = row.Scan(&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &project, nil
}
