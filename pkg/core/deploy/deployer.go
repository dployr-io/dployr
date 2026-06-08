// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"context"
	"time"

	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"
)

type Deployer struct {
	config *shared.Config
	logger *shared.Logger
	store  store.DeploymentStore
	api    HandleDeployment
}

func NewDeployer(c *shared.Config, l *shared.Logger, s store.DeploymentStore, a HandleDeployment) *Deployer {
	return &Deployer{
		config: c,
		logger: l,
		store:  s,
		api:    a,
	}
}

type DeployRequest struct {
	Name        string             `json:"name" validate:"required"`
	Description string             `json:"description,omitempty"`
	UserId      string             `json:"user_id" validate:"required"`
	Type        string             `json:"type" validate:"required,oneof= static web worker job"`
	Source      string             `json:"source" validate:"required,oneof=remote image"`
	Runtime     string             `json:"runtime" validate:"required,oneof=static golang php python nodejs ruby dotnet java"`
	Version     string             `json:"version,omitempty"`
	RunCmd      string             `json:"run_cmd,omitempty"`
	BuildCmd    string             `json:"build_cmd,omitempty"`
	Port        int                `json:"port,omitempty"`
	WorkingDir  string             `json:"working_dir,omitempty"`
	StaticDir   string             `json:"static_dir,omitempty"`
	Image       string             `json:"image,omitempty"`
	EnvVars     map[string]any     `json:"env_vars,omitempty"`
	Secrets     map[string]any     `json:"secrets,omitempty"`
	Remote      store.RemoteObj    `json:"remote,omitempty"`
	Domain      string             `json:"domain,omitempty"`
	HealthCheck string             `json:"health_check,omitempty"`
}

func (dr *DeployRequest) GetRuntimeObj() store.RuntimeObj {
	return store.RuntimeObj{
		Type:    store.Runtime(dr.Runtime),
		Version: dr.Version,
	}
}

type DeployResponse struct {
	Success   bool      `json:"success"`
	ID        string    `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type BuildRequest struct {
	DeployRequest
	CallbackInstance string `json:"callback_instance"`
}

type PublishRequest struct {
	Image   string        `json:"image"`
	Payload DeployRequest `json:"payload"`
}

type BuildResponse struct {
	Image string `json:"image"`
}

type HandleDeployment interface {
	Deploy(ctx context.Context, req *DeployRequest) (*DeployResponse, error)
	Build(ctx context.Context, req *BuildRequest) (*BuildResponse, error)
	Publish(ctx context.Context, req *PublishRequest) (*DeployResponse, error)
	GetDeployment(ctx context.Context, id string) (*store.Deployment, error)
	ListDeployments(ctx context.Context, id string, limit, offset int) ([]*store.Deployment, error)
	UpdateDeploymentStatus(ctx context.Context, id string, status store.Status) error
}
