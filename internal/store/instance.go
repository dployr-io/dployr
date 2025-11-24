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
        SELECT id, bootstrap_token, access_token, instance_id, registered_at, last_installed_at
        FROM instance LIMIT 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var inst store.Instance
		var registeredAtUnix, lastInstalledAtUnix int64
		var bootstrap, access sql.NullString

		if err := rows.Scan(&inst.ID, &bootstrap, &access, &inst.InstanceID, &registeredAtUnix, &lastInstalledAtUnix); err != nil {
			return nil, err
		}

		if bootstrap.Valid {
			inst.BootstrapToken = bootstrap.String
		}
		if access.Valid {
			inst.AccessToken = access.String
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

func (s *InstanceStore) SetBootstrapToken(ctx context.Context, token string) error {
	now := time.Now().Unix()
	res, err := s.db.ExecContext(ctx, `
		UPDATE instance SET bootstrap_token = ?, last_installed_at = ?`,
		token,
		now,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO instance (id, bootstrap_token, access_token, instance_id, issuer, audience, registered_at, last_installed_at)
			VALUES (?, ?, '', '', '', '', ?, ?)`,
			"instance",
			token,
			"",
			now,
			now,
		)
	}
	return err
}

func (s *InstanceStore) GetBootstrapToken(ctx context.Context) (string, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT bootstrap_token
		FROM instance LIMIT 1`)

	var token string
	if err := row.Scan(&token); err != nil {
		return "", err
	}

	return token, nil
}

func (s *InstanceStore) SetAccessToken(ctx context.Context, token string) error {
	now := time.Now().Unix()
	res, err := s.db.ExecContext(ctx, `
		UPDATE instance SET access_token = ?, last_installed_at = ?`,
		token,
		now,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO instance (id, bootstrap_token, access_token, instance_id, issuer, audience, registered_at, last_installed_at)
			VALUES (?, '', ?, '', '', ?, ?)`,
			"instance",
			token,
			now,
			now,
		)
	}
	return err
}

func (s *InstanceStore) GetAccessToken(ctx context.Context) (string, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT access_token
		FROM instance LIMIT 1`)

	var token sql.NullString
	if err := row.Scan(&token); err != nil {
		return "", err
	}

	if !token.Valid {
		return "", nil
	}

	return token.String, nil
}

func (s *InstanceStore) RegisterInstance(ctx context.Context, i *store.Instance) error {
	now := time.Now().Unix()
	res, err := s.db.ExecContext(ctx, `
		UPDATE instance SET instance_id = ?, issuer = ?, audience = ?, last_installed_at = ?`,
		i.InstanceID,
		i.Issuer,
		i.Audience,
		now,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO instance (id, bootstrap_token, access_token, instance_id, issuer, audience, registered_at, last_installed_at)
			VALUES (?, '', '', ?, ?, ?, ?)`,
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
