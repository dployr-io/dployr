// queue_manager.go - Updated with SQLite driver
package queue

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riversqlite"
	"github.com/riverqueue/river/rivertype"

	"dployr.io/pkg/jobs"
	"dployr.io/pkg/logger"
	"dployr.io/pkg/repository"
)

var (
	ErrInvalidJobData = errors.New("invalid job data format")
)

type QueueManager struct {
	client *river.Client[*sql.Tx]
	db     *sql.DB
	logger *logger.Logger
}

func NewQueueManager(r *repository.Event, logger *logger.Logger) (*QueueManager, error) {
	// Open SQLite database with WAL mode for better concurrency
	db, err := sql.Open("sqlite", "file:./data.sqlite3?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}

	// Configure connection pool for SQLite
	db.SetMaxOpenConns(1) // SQLite works best with single connection
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(1 * time.Hour)

	// Create worker registry
	workers := river.NewWorkers()

	// Register all workers using the simplified approach
	if err := jobs.RegisterWorkers(workers, r, logger); err != nil {
		return nil, fmt.Errorf("failed to register workers: %w", err)
	}

	// Create river client with SQLite driver
	riverClient, err := river.NewClient(riversqlite.New(db), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers: workers,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create river client: %w", err)
	}

	return &QueueManager{
		client: riverClient,
		db:     db,
		logger: logger,
	}, nil
}

// Helper methods for inserting specific job types
func (qm *QueueManager) InsertCreateProjectJob(ctx context.Context, tx *sql.Tx, userID string, data map[string]interface{}) (*rivertype.JobInsertResult, error) {
	args := jobs.NewCreateProjectArgs(userID, data)
	return qm.client.InsertTx(ctx, tx, args, nil)
}

func (qm *QueueManager) InsertBuildProjectJob(ctx context.Context, tx *sql.Tx, userID string, data map[string]interface{}) (*rivertype.JobInsertResult, error) {
	args := jobs.NewBuildProjectArgs(userID, data)
	return qm.client.InsertTx(ctx, tx, args, nil)
}

func (qm *QueueManager) InsertDeployProjectJob(ctx context.Context, tx *sql.Tx, userID string, data map[string]interface{}) (*rivertype.JobInsertResult, error) {
	args := jobs.NewDeployProjectArgs(userID, data)
	return qm.client.InsertTx(ctx, tx, args, nil)
}

func (qm *QueueManager) GetClient() *river.Client[*sql.Tx] {
	return qm.client
}

func (qm *QueueManager) Start(ctx context.Context) error {
	return qm.client.Start(ctx)
}

func (qm *QueueManager) Stop(ctx context.Context) error {
	return qm.client.Stop(ctx)
}

func (qm *QueueManager) Close() {
	qm.db.Close()
}

func (qm *QueueManager) GetDB() *sql.DB {
	return qm.db
}
