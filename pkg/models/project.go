package models

import (
	"time"
)

// States
const (
	Pending   = "pending"
	Building  = "building"
	Deploying = "deploying"
	Success   = "success"
	Failed    = "failed"
)

// Phases
const (
	Setup    = "setup"
	Build    = "build"
	Deploy   = "deploy"
	SSL      = "ssl"
	Complete = "complete"
)

type Project struct {
	ID            string             `db:"id" json:"id"`
	Name          string             `db:"name" json:"name"`
	GitRepo       *string            `db:"git_repo" json:"git_repo"`
	Domain        *string            `db:"domain" json:"domain"`
	Provider      *string            `db:"provider" json:"provider"`
	Environment   *map[string]string `db:"environment" json:"environment,omitempty"`
	LastDeployed  *time.Time         `db:"last_deployed" json:"last_deployed,omitempty"`
	DeploymentURL *string            `db:"deployment_url" json:"deployment_url,omitempty"`
	Status        string             `db:"status" json:"status"`
	HostConfigs   *map[string]any    `db:"host_configs" json:"host_configs,omitempty"`
	CreatedAt     time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `db:"updated_at" json:"updated_at"`
}

type SetupUserData struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}
