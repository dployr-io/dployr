package deploy

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"dployr/pkg/core/deploy"
	"dployr/pkg/shared"
	"dployr/pkg/store"

	"github.com/oklog/ulid/v2"
)

type Dispatcher interface {
	Submit(id string)
}

type Deployer struct {
	cfg       *shared.Config
	logger       *slog.Logger
	store        store.DeploymentStore
	job			 Dispatcher
}

// New creates a new Deployer instance
func New(c *shared.Config, l *slog.Logger, s store.DeploymentStore, j Dispatcher) *Deployer {
	return &Deployer{
		cfg:       c,
		logger:       l,
		store:        s,
		job: 	    	j,
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
		Cfg: store.Config{
			Name:       req.Name,
			Desc:       req.Description,
			Runtime:    req.Runtime,
			RunCmd:     req.RunCmd,
			BuildCmd:   req.BuildCmd,
			Port:       req.Port,
			WorkingDir: req.WorkingDir,
			StaticDir:  req.StaticDir,
			Image:      req.Image,
			EnvVars:    req.EnvVars,
			Remote:     req.Remote,
			Source:     req.Source,
		},
		SaveSpec:  req.SaveSpec,
		UserId:    &user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := d.store.CreateDeployment(ctx, deployment); err != nil {
		msg := fmt.Sprintf("failed to create deployment: %s", err)
		d.logger.With("request_id", requestID, "trace_id", traceID).Error(msg)
		return nil, fmt.Errorf("%s", msg)
	}

	// TODO: Validate deployment with JSON shema

	d.logger.With(
		"user_id", user.ID,
		"request_id", requestID,
		"trace_id", traceID,
	).Info("created new deployment", "deployment_id", deployment.ID)

	d.job.Submit(deployment.ID)

	return &deploy.DeployResponse{
		ID:        deployment.ID,
		Name:      deployment.Cfg.Name,
		Success:   true,
		CreatedAt: deployment.CreatedAt,
	}, nil
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
		msg := fmt.Sprintf("connot modify state after failure or completion: %s", err)
		d.logger.With("request_id", requestID).Error(msg)
		return fmt.Errorf("%s", msg)
	}

	return d.store.UpdateDeploymentStatus(ctx, id, string(status))
}
