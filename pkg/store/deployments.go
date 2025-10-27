package store

import (
	"context"
	"time"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusFailed     Status = "failed"
	StatusCompleted  Status = "completed"
)

type RuntimeObj struct {
	Type    Runtime `json:"type" db:"type" validate:"required,oneof=static go php python node-js ruby dotnet java docker k3s custom"`
	Version string  `json:"version,omitempty" db:"version"`
}

type RemoteObj struct {
	Url        string `json:"url" db:"url"`
	Branch     string `json:"branch" db:"branch"`
	CommitHash string `json:"commit_hash" db:"commit_hash"`
}

type Config struct {
	Name       string            `json:"name" db:"name"`
	Desc       string            `json:"description" db:"description"`
	Source     string            `json:"source" db:"source"`
	Runtime    RuntimeObj        `json:"runtime" db:"runtime"`
	Remote     RemoteObj         `json:"remote" db:"remote"`
	RunCmd     string            `json:"run_cmd,omitempty" db:"run_cmd"`
	BuildCmd   string            `json:"build_cmd,omitempty" db:"build_cmd"`
	Port       int               `json:"port" db:"port"`
	WorkingDir string            `json:"working_dir,omitempty" db:"working_dir"`
	StaticDir  string            `json:"static_dir,omitempty" db:"static_dir"`
	Image      string            `json:"image,omitempty" db:"image"`
	EnvVars    map[string]string `json:"env_vars,omitempty" db:"env_vars"`
	Status     string            `json:"status" db:"status"`
	ProjectID  *string           `json:"project_id,omitempty" db:"project_id"`
}

type Deployment struct {
	ID        string    `json:"id" db:"id"`
	UserId    *string   `json:"user_id,omitempty" db:"user_id"`
	Cfg       Config    `json:"config" db:"config"`
	Status    Status    `json:"status" db:"status"`
	SaveSpec  bool      `json:"save_spec" db:"save_spec"`
	Metadata  string    `json:"metadata" db:"metadata"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type DeploymentStore interface {
	CreateDeployment(ctx context.Context, d *Deployment) error
	GetDeployment(ctx context.Context, id string) (*Deployment, error)
	ListDeployments(ctx context.Context, limit, offset int) ([]*Deployment, error)
	UpdateDeploymentStatus(ctx context.Context, id string, status string) error
}
