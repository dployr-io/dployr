package jobs

import (
	"context"
	"time"

	"dployr.io/pkg/logger"
	"dployr.io/pkg/models"
	"dployr.io/pkg/repository"
)

// BaseJobArgs shared by all job types
type BaseJobArgs struct {
    UserID    string                 `json:"user_id"`
    JobID     string                 `json:"job_id"`
    Data      map[string]interface{} `json:"data"`
    CreatedAt time.Time              `json:"created_at"`
}

// Separate args types (River requires this)
type CreateProjectArgs struct {
    BaseJobArgs
}
func (CreateProjectArgs) Kind() string { return "create_project" }

type BuildProjectArgs struct {
    BaseJobArgs
}
func (BuildProjectArgs) Kind() string { return "build_project" }

type DeployProjectArgs struct {
    BaseJobArgs
}
func (DeployProjectArgs) Kind() string { return "deploy_project" }

// Shared handler logic
func handleCreateProject(ctx context.Context, args BaseJobArgs, r *repository.Event, log *logger.Logger) error {
    log.Info(ctx, r, models.Setup, "Setting up project for user " + args.UserID)
    return nil
}

func handleBuildProject(ctx context.Context, args BaseJobArgs, r *repository.Event, log *logger.Logger) error {
    log.Info(ctx, r, models.Build, "Building project for user " + args.UserID)
    return nil
}

func handleDeployProject(ctx context.Context, args BaseJobArgs, r *repository.Event, log *logger.Logger) error {
    log.Info(ctx, r, models.Deploy, "Deploying project for user " + args.UserID)
    return nil
}

// Helper functions
func NewCreateProjectArgs(userID, jobID string, data map[string]interface{}) CreateProjectArgs {
    return CreateProjectArgs{BaseJobArgs{userID, jobID, data, time.Now()}}
}

func NewBuildProjectArgs(userID, jobID string, data map[string]interface{}) BuildProjectArgs {
    return BuildProjectArgs{BaseJobArgs{userID, jobID, data, time.Now()}}
}

func NewDeployProjectArgs(userID, jobID string, data map[string]interface{}) DeployProjectArgs {
    return DeployProjectArgs{BaseJobArgs{userID, jobID, data, time.Now()}}
}