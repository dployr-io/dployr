// queue_manager.go - Updated with simplified worker registration
package queue

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"

	"dployr.io/pkg/config"
	"dployr.io/pkg/jobs"
	"dployr.io/pkg/logger"
	"dployr.io/pkg/repository"
)

var (
	ErrInvalidJobData = errors.New("invalid job data format")
)

type QueueManager struct {
	client *river.Client[pgx.Tx]
	pool   *pgxpool.Pool
	logger *logger.Logger
}

func NewQueueManager(r *repository.Event, logger *logger.Logger) (*QueueManager, error) {
	poolConfig, err := pgxpool.ParseConfig(config.GetDSN("5432"))
	if err != nil {
		return nil, err
	}

	// Configure connection pool for better performance
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Create worker registry
	workers := river.NewWorkers()

	// Register all workers using the simplified approach
	if err := jobs.RegisterWorkers(workers, r, logger); err != nil {
		return nil, fmt.Errorf("failed to register workers: %w", err)
	}

	// Create river client
	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
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
		pool:   pool,
		logger: logger,
	}, nil
}

// Helper methods for inserting specific job types
func (qm *QueueManager) InsertCreateProjectJob(ctx context.Context, tx pgx.Tx, userID, jobID string, data map[string]interface{}) (*rivertype.JobInsertResult, error) {
    args := jobs.NewCreateProjectArgs(userID, jobID, data)
	log.Println("Inserting create project job for user " + userID)
    return qm.client.InsertTx(ctx, tx, args, nil)
}

func (qm *QueueManager) InsertBuildProjectJob(ctx context.Context, tx pgx.Tx, userID, jobID string, data map[string]interface{}) (*rivertype.JobInsertResult, error) {
    args := jobs.NewBuildProjectArgs(userID, jobID, data)
    return qm.client.InsertTx(ctx, tx, args, nil)
}

func (qm *QueueManager) InsertDeployProjectJob(ctx context.Context, tx pgx.Tx, userID, jobID string, data map[string]interface{}) (*rivertype.JobInsertResult, error) {
    args := jobs.NewDeployProjectArgs(userID, jobID, data)
    return qm.client.InsertTx(ctx, tx, args, nil)
}

func (qm *QueueManager) GetClient() *river.Client[pgx.Tx] {
	return qm.client
}

func (qm *QueueManager) Start(ctx context.Context) error {
	return qm.client.Start(ctx)
}

func (qm *QueueManager) Stop(ctx context.Context) error {
	return qm.client.Stop(ctx)
}

func (qm *QueueManager) Close() {
	qm.pool.Close()
}

func (qm *QueueManager) GetPool() *pgxpool.Pool {
	return qm.pool
}
