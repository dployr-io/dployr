package store

import (
	"context"
	"time"
)

type Runtime string

const (
	RuntimeStatic Runtime = "static"
	RuntimeGo     Runtime = "golang"
	RuntimePHP    Runtime = "php"
	RuntimePython Runtime = "python"
	RuntimeNodeJS Runtime = "nodejs"
	RuntimeRuby   Runtime = "ruby"
	RuntimeDotnet Runtime = "dotnet"
	RuntimeJava   Runtime = "java"
	RuntimeDocker Runtime = "docker"
	RuntimeK3S    Runtime = "k3s"
	RuntimeCustom Runtime = "custom"
)

type Service struct {
	ID             string         `json:"id" db:"id"`
	Name           string         `json:"name" db:"name"`
	Description    string         `json:"description" db:"description"`
	Source         string         `json:"source" db:"source"`
	Runtime        Runtime        `json:"runtime" db:"runtime"`
	RuntimeVersion string         `json:"runtime_version,omitempty" db:"runtime_version"`
	RunCmd         string         `json:"run_cmd,omitempty" db:"run_cmd"`
	BuildCmd       string         `json:"build_cmd,omitempty" db:"build_cmd"`
	Port           int            `json:"port" db:"port"`
	WorkingDir     string         `json:"working_dir,omitempty" db:"working_dir"`
	StaticDir      string         `json:"static_dir,omitempty" db:"static_dir"`
	Image          string         `json:"image,omitempty" db:"image"`
	EnvVars        string         `json:"env_vars,omitempty" db:"env_vars"`
	Status         string         `json:"status" db:"status"`
	ProjectID      *string `json:"project_id,omitempty" db:"project_id"`
	RemoteID       *string `json:"remote_id,omitempty" db:"remote_id"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}

type ServiceStore interface {
	CreateService(ctx context.Context, svc *Service) error
	GetService(ctx context.Context, id string) (*Service, error)
}
