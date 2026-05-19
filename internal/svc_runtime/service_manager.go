// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package svc_runtime

import (
	"fmt"
	"runtime"
)

// ServiceManager defines the interface for service management across platforms
type ServiceManager interface {
	Status(name string) (string, error)
	HealthStatus(name string) (string, error)
	Install(name, desc, runCmd, workDir string, envVars map[string]string) error
	Start(name string) error
	Stop(name string) error
	Remove(name string) error
	Ice(name string) error
}

func SvcRuntime() (ServiceManager, error) {
	cli, err := NewDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker daemon: %w", err)
	}
	ds := DockerService{cli: cli}
	switch runtime.GOOS {
	case "linux":
		return &SystemdManager{DockerService: ds}, nil
	case "darwin":
		return &LaunchdManager{DockerService: ds}, nil
	case "windows":
		return &NSSMManager{DockerService: ds}, nil
	default:
		return nil, fmt.Errorf("no compatible runtime was detected")
	}
}

func GetSvcMgrName() string {
	switch runtime.GOOS {
	case "linux":
		return "systemd on Linux"
	case "darwin":
		return "launchd on macOS"
	case "windows":
		return "NSSM on Windows"
	case "docker":
		return "docker"
	default:
		return "systemd on Linux, launchd on macOS, or NSSM on Windows or docker"
	}
}
