package jobs

import (
	"context"

	"github.com/riverqueue/river"

	"dployr.io/pkg/logger"
	"dployr.io/pkg/repository"
)

func RegisterWorkers(workers *river.Workers, r *repository.EventRepo, log *logger.Logger) error {
	river.AddWorker(workers, river.WorkFunc(func(ctx context.Context, job *river.Job[CreateProjectArgs]) error {
		return handleCreateProject(ctx, job, r, log)
	}))

	river.AddWorker(workers, river.WorkFunc(func(ctx context.Context, job *river.Job[BuildProjectArgs]) error {
		return handleBuildProject(ctx, job, r, log)
	}))

	river.AddWorker(workers, river.WorkFunc(func(ctx context.Context, job *river.Job[DeployProjectArgs]) error {
		return handleDeployProject(ctx, job, r, log)
	}))

	return nil
}
