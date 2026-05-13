// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"context"
	"fmt"
	"time"

	"github.com/dployr-io/dployr/pkg/core/deploy"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"

	"github.com/oklog/ulid/v2"
)

type Dispatcher interface {
	Submit(id string)
}

type Deployer struct {
	cfg    *shared.Config
	logger *shared.Logger
	store  store.DeploymentStore
	job    Dispatcher
}

// Init creates a new Deployer instance
func Init(c *shared.Config, l *shared.Logger, s store.DeploymentStore, j Dispatcher) *Deployer {
	return &Deployer{
		cfg:    c,
		logger: l,
		store:  s,
		job:    j,
	}
}

func (d *Deployer) Deploy(ctx context.Context, req *deploy.DeployRequest) (*deploy.DeployResponse, error) {
	requestID, err := shared.TraceFromContext(ctx)
	if err != nil {
		d.logger.Error("failed to extract request id from context")
		return nil, fmt.Errorf("failed to extract request id from context: %w", err)
	}

	traceID, err := shared.TraceFromContext(ctx)
	if err != nil {
		d.logger.With("request_id", requestID).Error("failed to extract trace id from context")
		return nil, fmt.Errorf("failed to extract trace id from context: %w", err)
	}

	user, err := shared.UserFromContext(ctx)
	if err != nil {
		d.logger.With("request_id", requestID, "trace_id", traceID).Error("unauthenticated deployment attempt")
		return nil, fmt.Errorf("unauthenticated deployment attempt")
	}

	deployment := &store.Deployment{
		ID:     ulid.Make().String(),
		Status: store.StatusPending,
		Name:   req.Name,
		Blueprint: store.Blueprint{
			Name:        req.Name,
			Desc:        req.Description,
			Type:        store.ServiceType(req.Type),
			Runtime:     req.GetRuntimeObj(),
			RunCmd:      req.RunCmd,
			BuildCmd:    req.BuildCmd,
			Port:        req.Port,
			WorkingDir:  req.WorkingDir,
			StaticDir:   req.StaticDir,
			Image:       req.Image,
			EnvVars:     shared.ConvertMapToStrings(req.EnvVars),
			Secrets:     shared.ConvertMapToStrings(req.Secrets),
			Remote:      req.Remote,
			Source:      store.Source(req.Source),
			HealthCheck: req.HealthCheck,
		},
		UserId:    &user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := d.store.UpsertDeployment(ctx, deployment); err != nil {
		msg := fmt.Sprintf("failed to upsert deployment: %s", err)
		d.logger.With("request_id", requestID, "trace_id", traceID).Error(msg)
		return nil, fmt.Errorf("%s", msg)
	}

	d.logger.With(
		"user_id", user.ID,
		"request_id", requestID,
		"trace_id", traceID,
	).Info("upserted deployment", "deployment_id", deployment.ID)

	d.job.Submit(deployment.ID)

	return &deploy.DeployResponse{
		ID:        deployment.ID,
		Name:      deployment.Blueprint.Name,
		Success:   true,
		CreatedAt: deployment.CreatedAt,
	}, nil
}

func (d *Deployer) Build(ctx context.Context, req *deploy.BuildRequest) (*deploy.BuildResponse, error) {
	workDir, err := SetupDir(req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to setup working directory: %w", err)
	}

	if err := CloneRepo(req.Remote, workDir, req.WorkingDir, d.cfg); err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	image, err := BuildImage(req.Name, workDir, d.cfg)
	if err != nil {
		return nil, fmt.Errorf("build failed: %w", err)
	}

	return &deploy.BuildResponse{Image: image}, nil
}

func (d *Deployer) Publish(ctx context.Context, req *deploy.PublishRequest) (*deploy.DeployResponse, error) {
	deployReq := req.Payload
	deployReq.Source = string(store.SourceImage)
	deployReq.Image = req.Image
	return d.Deploy(ctx, &deployReq)
}

func (d *Deployer) GetDeployment(ctx context.Context, id string) (*store.Deployment, error) {
	return d.store.GetDeployment(ctx, id)
}

func (d *Deployer) ListDeployments(ctx context.Context, userID string, limit, offset int) ([]*store.Deployment, error) {
	return d.store.ListDeployments(ctx, limit, offset)
}

func (d *Deployer) UpdateDeploymentStatus(ctx context.Context, id string, status store.Status) error {
	requestID, err := shared.TraceFromContext(ctx)
	if err != nil {
		msg := fmt.Sprintf("failed to extract request id from context: %s", err)
		d.logger.Error(msg)
		return fmt.Errorf("%s", msg)
	}

	_, err = shared.UserFromContext(ctx)
	if err != nil {
		msg := fmt.Sprintf("unauthenticated deployment attempt: %s", err)
		d.logger.With("request_id", requestID).Error(msg)
		return fmt.Errorf("%s", msg)
	}

	if status == store.StatusCompleted || status == store.StatusFailed {
		msg := fmt.Sprintf("cannot modify state after failure or completion: %s", err)
		d.logger.With("request_id", requestID).Error(msg)
		return fmt.Errorf("%s", msg)
	}

	return d.store.UpdateDeploymentStatus(ctx, id, string(status))
}
