package jobs

import (
	"context"
	"fmt"

	"github.com/riverqueue/river"

	"dployr.io/pkg/logger"
	"dployr.io/pkg/models"
	"dployr.io/pkg/repository"
)

// Base job arguments that all jobs share
type BaseJobArgs struct {
	UserID string                 `json:"user_id"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type CreateProjectArgs struct {
	BaseJobArgs
}

type BuildProjectArgs struct {
	BaseJobArgs
}

type DeployProjectArgs struct {
	BaseJobArgs
}

// Kind methods for River interface
func (CreateProjectArgs) Kind() string { return "create_project" }
func (BuildProjectArgs) Kind() string  { return "build_project" }
func (DeployProjectArgs) Kind() string { return "deploy_project" }

// Shared handler logic using the full job object to access River's metadata
func handleCreateProject(ctx context.Context, job *river.Job[CreateProjectArgs], r *repository.Event, log *logger.Logger) error {
	log.Info(ctx, r, models.Setup, fmt.Sprintf("Setting up project for user %s (job created at %v)", job.Args.UserID, job.CreatedAt))
	return nil
}

func handleBuildProject(ctx context.Context, job *river.Job[BuildProjectArgs], r *repository.Event, log *logger.Logger) error {
	log.Info(ctx, r, models.Build, fmt.Sprintf("Building project for user %s (job created at %v)", job.Args.UserID, job.CreatedAt))
	return nil
}

func handleDeployProject(ctx context.Context, job *river.Job[DeployProjectArgs], r *repository.Event, log *logger.Logger) error {
	log.Info(ctx, r, models.Deploy, fmt.Sprintf("Deploying project for user %s (job created at %v)", job.Args.UserID, job.CreatedAt))
	return nil
}

// Helper functions - using composition
func newBaseJobArgs(userID string, data map[string]interface{}) BaseJobArgs {
	return BaseJobArgs{UserID: userID, Data: data}
}

func NewCreateProjectArgs(userID string, data map[string]interface{}) CreateProjectArgs {
	return CreateProjectArgs{BaseJobArgs: newBaseJobArgs(userID, data)}
}

func NewBuildProjectArgs(userID string, data map[string]interface{}) BuildProjectArgs {
	return BuildProjectArgs{BaseJobArgs: newBaseJobArgs(userID, data)}
}

func NewDeployProjectArgs(userID string, data map[string]interface{}) DeployProjectArgs {
	return DeployProjectArgs{BaseJobArgs: newBaseJobArgs(userID, data)}
}
