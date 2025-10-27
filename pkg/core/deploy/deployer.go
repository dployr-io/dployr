package deploy

import (
	"context"
	"log/slog"
	"time"

	"dployr/pkg/shared"
	"dployr/pkg/store"
)

type Deployer struct {
	config *shared.Config
	logger *slog.Logger
	store  store.DeploymentStore
	api HandleDeployment
}

func NewDeployer(c *shared.Config, l *slog.Logger, s store.DeploymentStore, a HandleDeployment) *Deployer {
	return &Deployer{
		config: c,
		logger: l,
		store:  s,
		api: a,
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

type HandleDeployment interface {
	Deploy(ctx context.Context, req *DeployRequest) (*DeployResponse, error)
	GetDeployment(ctx context.Context, id string) (*store.Deployment, error)
	ListDeployments(ctx context.Context, userID string, limit, offset int) ([]*store.Deployment, error)
	UpdateDeploymentStatus(ctx context.Context, id string, status store.Status) error
}
