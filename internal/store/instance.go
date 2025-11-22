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
        FROM instance LIMIT 1`)
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
		UPDATE instance SET last_installed_at = ?`,
		time.Now().Unix(),
	)
	return err
}

func (s *InstanceStore) SetToken(ctx context.Context, token string) error {
	now := time.Now().Unix()
	res, err := s.db.ExecContext(ctx, `
		UPDATE instance SET token = ?, last_installed_at = ?`,
		token,
		now,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO instance (id, token, instance_id, issuer, audience, registered_at, last_installed_at)
			VALUES (?, ?, '', '', '', ?, ?)`,
			"instance",
			token,
			now,
			now,
		)
	}
	return err
}

func (s *InstanceStore) GetToken(ctx context.Context) (string, error) {
	row := s.db.QueryRowContext(ctx, `
        SELECT token
        FROM instance LIMIT 1`)

	var token string
	if err := row.Scan(&token); err != nil {
		return "", err
	}

	return token, nil
}

func (s *InstanceStore) RegisterInstance(ctx context.Context, i *store.Instance) error {
	now := time.Now().Unix()
	res, err := s.db.ExecContext(ctx, `
		UPDATE instance SET instance_id = ?, issuer = ?, audience = ?, registered_at = ?, last_installed_at = ?`,
		i.InstanceID,
		i.Issuer,
		i.Audience,
		now,
		now,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO instance (id, token, instance_id, issuer, audience, registered_at, last_installed_at)
			VALUES (?, '', ?, ?, ?, ?, ?)`,
			"instance",
			i.InstanceID,
			i.Issuer,
			i.Audience,
			now,
			now,
		)
	}
	return err
}
