package store

import (
	"database/sql"
	"fmt"
)

type BootstrapTokenStore struct {
	db *sql.DB
}

func (s *BootstrapTokenStore) Create(instanceId, nonce string) error {
	_, err := s.db.Exec(
		`INSERT INTO bootstrap_tokens (instance_id, nonce) VALUES (?, ?)`,
		instanceId, nonce,
	)
	return err
}

func (s *BootstrapTokenStore) MarkUsed(nonce string) error {
	result, err := s.db.Exec(
		`UPDATE bootstrap_tokens SET used_at = CURRENT_TIMESTAMP 
         WHERE nonce = ? AND used_at IS NULL`,
		nonce,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("token already used or not found")
	}
	return nil
}
