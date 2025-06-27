package jobs

import (
	"context"

	"github.com/riverqueue/river"

	"dployr.io/pkg/logger"
	"dployr.io/pkg/repository"
)

func RegisterWorkers(workers *river.Workers, r *repository.Event, log *logger.Logger) error {
    river.AddWorker(workers, river.WorkFunc(func(ctx context.Context, job *river.Job[CreateProjectArgs]) error {
        return handleCreateProject(ctx, job.Args.BaseJobArgs, r, log)
    }))

    river.AddWorker(workers, river.WorkFunc(func(ctx context.Context, job *river.Job[BuildProjectArgs]) error {
        return handleBuildProject(ctx, job.Args.BaseJobArgs, r, log)
    }))

    river.AddWorker(workers, river.WorkFunc(func(ctx context.Context, job *river.Job[DeployProjectArgs]) error {
        return handleDeployProject(ctx, job.Args.BaseJobArgs, r, log)
    }))

    return nil
}