package models

import (
	"encoding/json"
	"time"
)

// States
const (
	Pending = "setup"
	Building = "building"
	Deploying = "deploying"
	Success = "success"
	Failed = "failed"
)

// Phases
const (
    Setup = "setup"
	Build = "build"
	Deploy = "deploy"
	SSL = "ssl"
	Complete = "complete"
)

type Project struct {
    Id            string            `db:"id" json:"id"`
    Name          string            `db:"name" json:"name"`
    GitRepo       string            `db:"git_repo" json:"git_repo"`
    Domain        *string            `db:"domain" json:"domain,omitempty"`
    Environment   *json.RawMessage   `db:"environment" json:"environment,omitempty"`
    CreatedAt     time.Time         `db:"created_at" json:"created_at"`
    UpdatedAt     time.Time         `db:"updated_at" json:"updated_at"`
    DeploymentURL *string            `db:"deployment_url" json:"deployment_url,omitempty"`
    LastDeployed  *time.Time        `db:"last_deployed" json:"last_deployed,omitempty"`
	Status        string            `db:"status" json:"status,omitempty"`
    HostConfigs   *json.RawMessage   `db:"host_configs" json:"host_configs,omitempty"`
}
