package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/dployr-io/dployr/pkg/store"
)

// TaskResultStore implements store.TaskResultStore using SQLite.
type TaskResultStore struct {
	db *sql.DB
}

func NewTaskResultStore(db *sql.DB) *TaskResultStore {
	return &TaskResultStore{db: db}
}

func (s *TaskResultStore) ListUnsent(ctx context.Context) ([]*store.TaskResult, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, status, result, error
		FROM task_results
		WHERE synced_at IS NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*store.TaskResult
	for rows.Next() {
		var r store.TaskResult
		var resultJSON []byte
		if err := rows.Scan(&r.ID, &r.Status, &resultJSON, &r.Error); err != nil {
			return nil, err
		}
		if len(resultJSON) > 0 {
			if err := json.Unmarshal(resultJSON, &r.Result); err != nil {
				// unmarshalable
			}
		}
		results = append(results, &r)
	}
	return results, rows.Err()
}

func (s *TaskResultStore) SaveResults(ctx context.Context, results []*store.TaskResult) error {
	if len(results) == 0 {
		return nil
	}

	stmt, err := s.db.PrepareContext(ctx, `
		INSERT INTO task_results (id, status, result, error, created_at)
		VALUES (?, ?, ?, ?, strftime('%s','now'))
		ON CONFLICT(id) DO UPDATE SET
			status=excluded.status,
			result=excluded.result,
			error=excluded.error`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range results {
		payload, err := json.Marshal(r.Result)
		if err != nil {
			payload = []byte("null")
		}
		if _, err := stmt.ExecContext(ctx, r.ID, r.Status, payload, r.Error); err != nil {
			return err
		}
	}
	return nil
}

func (s *TaskResultStore) MarkSynced(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids)+1)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := `
		UPDATE task_results
		SET synced_at = strftime('%s','now')
		WHERE id IN (` + strings.Join(placeholders, ",") + `)`

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}
