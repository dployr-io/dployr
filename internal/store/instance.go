package store

import (
	"context"
	"database/sql"
	"time"

	"dployr/pkg/store"
)

// InstanceStore provides DB-backed access to the single local instance record.
type InstanceStore struct {
	db *sql.DB
}

func NewInstanceStore(db *sql.DB) *InstanceStore {
	return &InstanceStore{db: db}
}

func (s *InstanceStore) GetInstance(ctx context.Context) (*store.Instance, error) {
	rows, err := s.db.QueryContext(ctx, `
        SELECT id, token, instance_id, registered_at, last_installed_at
        FROM instances LIMIT 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var inst store.Instance
		var registeredAtUnix, lastInstalledAtUnix int64

		if err := rows.Scan(&inst.ID, &inst.Token, &inst.InstanceID, &registeredAtUnix, &lastInstalledAtUnix); err != nil {
			return nil, err
		}

		inst.RegisteredAt = time.Unix(registeredAtUnix, 0)
		inst.LastInstalledAt = time.Unix(lastInstalledAtUnix, 0)

		return &inst, nil
	}

	return nil, nil
}

func (s *InstanceStore) UpdateLastInstalledAt(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE instances SET last_installed_at = ?`,
		time.Now().Unix(),
	)
	return err
}

func (s *InstanceStore) GetToken(ctx context.Context) (string, error) {
	row := s.db.QueryRowContext(ctx, `
        SELECT token
        FROM instances LIMIT 1`)

	var token string
	if err := row.Scan(&token); err != nil {
		return "", err
	}

	return token, nil
}

func (s *InstanceStore) RegisterInstance(ctx context.Context, i *store.Instance) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE instances SET instance_id = ?, token = ?, issuer = ?, audience = ?, registered_at = ?, last_installed_at = ?`,
		i.InstanceID,
		i.Token,
		i.Issuer,
		i.Audience,
		time.Now().Unix(),
		time.Now().Unix(),
	)
	return err
}
