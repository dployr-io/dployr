// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

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
	ID             string            `json:"id" db:"id"`
	Name           string            `json:"name" db:"name"`
	Description    string            `json:"description" db:"description"`
	Source         string            `json:"source" db:"source"`
	Runtime        Runtime           `json:"runtime" db:"runtime"`
	RuntimeVersion string            `json:"runtime_version,omitempty" db:"runtime_version"`
	RunCmd         string            `json:"run_cmd,omitempty" db:"run_cmd"`
	BuildCmd       string            `json:"build_cmd,omitempty" db:"build_cmd"`
	Port           int               `json:"port"`
	WorkingDir     string            `json:"working_dir,omitempty" db:"working_dir"`
	StaticDir      string            `json:"static_dir,omitempty" db:"static_dir"`
	Image          string            `json:"image,omitempty" db:"image"`
	EnvVars        map[string]string `json:"env_vars,omitempty"`
	Status         string            `json:"status"`
	Remote         string            `json:"remote,omitempty" db:"remote_url"`
	Branch         string            `json:"branch" db:"remote_branch"`
	CommitHash     string            `json:"commit_hash" db:"remote_commit_hash"`
	DeploymentId   string            `json:"-" db:"deployment_id"`
	Blueprint      *Blueprint        `json:"blueprint,omitempty"`
	CreatedAt      time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at" db:"updated_at"`
}

type ServiceStore interface {
	CreateService(ctx context.Context, svc *Service) (*Service, error)
	GetService(ctx context.Context, id string) (*Service, error)
	ListServices(ctx context.Context, limit, offset int) ([]*Service, error)
	UpdateService(ctx context.Context, svc *Service) error
}
