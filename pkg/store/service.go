// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"context"
	"time"
)

type NodeRole string

const (
	NodeRoleInstance NodeRole = "instance"
	NodeRoleBuild    NodeRole = "build"
)

type Runtime string

const (
	RuntimeGo     Runtime = "golang"
	RuntimePHP    Runtime = "php"
	RuntimePython Runtime = "python"
	RuntimeNodeJS Runtime = "nodejs"
	RuntimeRuby   Runtime = "ruby"
	RuntimeDotnet Runtime = "dotnet"
	RuntimeJava   Runtime = "java"
)

type ServiceType string

const (
	TypeStatic ServiceType = "static"
	TypeWorker ServiceType = "worker"
	TypeWeb    ServiceType = "web"
	TypeJob    ServiceType = "job"
)

type Service struct {
	ID             string            `json:"id" db:"id"`
	Name           string            `json:"name" db:"name"`
	Description    string            `json:"description" db:"description"`
	Source         Source            `json:"source" db:"source"`
	Type           ServiceType       `json:"type" db:"type"`
	Runtime        Runtime           `json:"runtime" db:"runtime"`
	RuntimeVersion string            `json:"runtime_version,omitempty" db:"runtime_version"`
	RunCmd         string            `json:"run_cmd,omitempty" db:"run_cmd"`
	BuildCmd       string            `json:"build_cmd,omitempty" db:"build_cmd"`
	Port           int               `json:"port"`
	WorkingDir     string            `json:"working_dir,omitempty" db:"working_dir"`
	StaticDir      string            `json:"static_dir,omitempty" db:"static_dir"`
	Image          string            `json:"image,omitempty" db:"image"`
	EnvVars        map[string]string `json:"env_vars,omitempty"`
	Secrets        map[string]string `json:"secrets,omitempty"`
	Remote         string            `json:"remote,omitempty" db:"remote_url"`
	Branch         string            `json:"branch" db:"remote_branch"`
	CommitHash     string            `json:"commit_hash" db:"remote_commit_hash"`
	DeploymentId   string            `json:"-" db:"deployment_id"`
	Blueprint      *Blueprint        `json:"blueprint,omitempty"`
	CreatedAt      time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at" db:"updated_at"`
}

type ServiceStore interface {
	GetService(ctx context.Context, name string) (*Service, error)
	ListServices(ctx context.Context, limit, offset int) ([]*Service, error)
	UpsertService(ctx context.Context, svc *Service) (*Service, error)
	DeleteService(ctx context.Context, name string) error
}
