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
}

func NewDeployer(config *shared.Config, logger *slog.Logger, store store.DeploymentStore) *Deployer {
	return &Deployer{
		config: config,
		logger: logger,
		store:  store,
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
	Port       	int              `json:"port,omitempty" validate:"required_unless=Runtime static docker k3s,omitempty,number"`
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
	traceID, err := shared.TraceFromContext(ctx)

	user, err := shared.UserFromContext(ctx)
	if err != nil {
		d.logger.With("request_id", requestID, "trace_id", traceID).Error("unauthenticated deployment attempt")
		return nil, fmt.Errorf(string(shared.BadRequest))
	}

	project, err := shared.ProjectFromContext(ctx)
	if err != nil {
		d.logger.With("request_id", requestID, "trace_id", traceID).Error("project_id not found in context")
		return nil, fmt.Errorf(string(shared.BadRequest))
	}

	deployment := &store.Deployment{
		ID:     ulid.Make().String(),
		Status: store.StatusPending,
		Config: store.Config{
			Name:        req.Name,
			Description: req.Description,
			Runtime:     req.Runtime,
			RunCmd:      req.RunCmd,
			BuildCmd:    req.BuildCmd,
			Port:        req.Port,
			WorkingDir:  req.WorkingDir,
			StaticDir:   req.StaticDir,
			Image:       req.Image,
			EnvVars:     req.EnvVars,
			Secrets:     req.Secrets,
			Remote:      req.Remote,
			Source:      req.Source,
		},
		SaveSpec: req.SaveSpec,
		UserId:  &user.ID,
	}

	if err := d.store.CreateDeployment(ctx, deployment); err != nil {
		return nil, fmt.Errorf(string(shared.RuntimeError))
	}

	// TODO: Validate deployment with JSON shema

	d.logger.With(
		"user_id", user.ID,
		"project_id", project.ID,
		"request_id", requestID,
		"trace_id", traceID,
	).Info("created new deployment", "deployment_id", deployment.ID)

	return &DeployResponse{
		ID:        deployment.ID,
		Name:      deployment.Config.Name,
		Success:   true,
		CreatedAt: deployment.CreatedAt,
	}, nil
}

func (d *Deployer) GetDeployment(ctx context.Context, id string) (*store.Deployment, error) {
	return d.store.GetDeploymentByID(ctx, id)
}

func (d *Deployer) ListDeployments(ctx context.Context, userID string, limit, offset int) ([]*store.Deployment, error) {
	return d.store.ListDeployments(ctx, limit, offset)
}

func (d *Deployer) UpdateDeploymentStatus(ctx context.Context, id string, status store.Status) error {
	requestID, err := shared.TraceFromContext(ctx)

	_, err = shared.UserFromContext(ctx)
	if err != nil {
		d.logger.With("request_id", requestID).Error("unauthenticated deployment attempt")
		return fmt.Errorf(string(shared.BadRequest))
	}

	_, err = shared.ProjectFromContext(ctx)
	if err != nil {
		d.logger.With("request_id", requestID).Error("project_id not found in context")
		return fmt.Errorf(string(shared.BadRequest))
	}

	if status == store.StatusCompleted || status == store.StatusFailed {
		d.logger.With("request_id", requestID).Error("connot modify state after failure or completion")
		return fmt.Errorf(string(shared.BadRequest))
	}

	return d.store.UpdateDeploymentStatus(ctx, id, string(status))
}
