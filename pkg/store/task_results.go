package store

import "context"

// TaskResult represents the outcome of executing a remote task.
type TaskResult struct {
	ID     string `db:"id"`
	Status string `db:"status"`
	Result any    `db:"result"`
	Error  string `db:"error"`
}

// TaskResultStore provides access to persisted task results.
type TaskResultStore interface {
	// ListUnsent returns all results that have not yet been marked as synced.
	ListUnsent(ctx context.Context) ([]*TaskResult, error)
	// SaveResults persists new task results as unsynced.
	SaveResults(ctx context.Context, results []*TaskResult) error
	// MarkSynced marks the given result IDs as synced.
	MarkSynced(ctx context.Context, ids []string) error
}
