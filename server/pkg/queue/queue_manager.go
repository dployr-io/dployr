package queue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"

	"dployr.io/pkg/config"
)

var (
	ErrInvalidJobData = errors.New("invalid job data format")
)

type QueueManager struct {
	client *river.Client[pgx.Tx]
	pool   *pgxpool.Pool
}

func NewQueueManager() (*QueueManager, error) {
	poolConfig, err := pgxpool.ParseConfig(config.GetDSN("5432"))
	if err != nil {
		return nil, err
	}

	// Configure connection pool for better performance
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	// Register all workers
	workers := river.NewWorkers()
	river.AddWorker(workers, &SetupUserWorker{})

	// Create River client
	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers:              workers,
		RescueStuckJobsAfter: 10 * time.Minute,
		JobTimeout:           10 * time.Minute,
	})

	if err != nil {
		return nil, err
	}

	return &QueueManager{
		client: client,
		pool:   pool,
	}, nil
}

func (qm *QueueManager) Start(ctx context.Context) error {
	return qm.client.Start(ctx)
}

func (qm *QueueManager) Stop(ctx context.Context) error {
	return qm.client.Stop(ctx)
}

// GetClient returns the River client for external use (e.g., River UI)
func (qm *QueueManager) GetClient() *river.Client[pgx.Tx] {
	return qm.client
}

// GetPool returns the pgxpool for external use (e.g., River UI)
func (qm *QueueManager) GetPool() *pgxpool.Pool {
	return qm.pool
}

// Transactional enqueue (recommended)
func (qm *QueueManager) EnqueueJobTx(ctx context.Context, tx pgx.Tx, jobArgs JobArgs, opts *river.InsertOpts) error {
	switch args := jobArgs.(type) {
	case SetupUserJobArgs:
		_, err := qm.client.InsertTx(ctx, tx, args, opts)
		return err
	default:
		return errors.New("unknown job type")
	}
}

// Helper functions for creating jobs
func NewJob(userID string, data interface{}) *BaseJobArgs {
	return &BaseJobArgs{
		JobID:     generateJobID(),
		UserID:    userID,
		Data:      data,
		CreatedAt: time.Now(),
	}
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}
