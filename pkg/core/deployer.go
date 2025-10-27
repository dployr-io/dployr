package core

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"dployr/pkg/shared"
	"dployr/pkg/store"

	"github.com/oklog/ulid/v2"
)

type Deployer struct {
	config *shared.Config
	logger *slog.Logger
	store  store.DeploymentStore
	worker *Worker
}

func NewDeployer(c *shared.Config, l *slog.Logger, s store.DeploymentStore, w *Worker) *Deployer {
	return &Deployer{
		config: c,
		logger: l,
		store:  s,
		worker: w,
	}
}

type DeployRequest struct {
	Name        string            `json:"name" validate:"required"`
	Description string            `json:"description,omitempty"`
	UserId      string            `json:"user_id" validate:"required"`
	Source      string            `json:"source" validate:"required,oneof=remote image"`
	Runtime     store.RuntimeObj  `json:"runtime" validate:"required"`
	Version     string            `json:"version,omitempty" validate:"omitempty"`
	RunCmd      string            `json:"run_cmd,omitempty" validate:"required_unless=Runtime static docker k3s,omitempty"`
	BuildCmd    string            `json:"build_cmd,omitempty" validate:"omitempty"`
	Port        int               `json:"port,omitempty" validate:"required_unless=Runtime static docker k3s,omitempty,number"`
	WorkingDir  string            `json:"working_dir,omitempty" validate:"omitempty"`
	StaticDir   string            `json:"static_dir,omitempty" validate:"omitempty"`
	Image       string            `json:"image,omitempty" validate:"omitempty"`
	SaveSpec    bool              `json:"save_spec,omitempty" validate:"omitempty"`
	EnvVars     map[string]string `json:"env_vars,omitempty" validate:"omitempty"`
	Secrets     map[string]string `json:"secrets,omitempty" validate:"omitempty"`
	Remote      store.RemoteObj   `json:"remote,omitempty" validate:"omitempty"`
	Domain      string            `json:"domain,omitempty" validate:"omitempty"`
	DNSProvider string            `json:"dns_provider,omitempty" validate:"omitempty"`
}

type DeployResponse struct {
	Success   bool      `json:"success"`
	ID        string    `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

func (d *Deployer) Deploy(ctx context.Context, req *DeployRequest) (*DeployResponse, error) {
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

	d.worker.Submit(deployment.ID)

	return &DeployResponse{
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
