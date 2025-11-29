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
	Install(name, desc, runCmd, workDir string, envVars map[string]string) error
	Start(name string) error
	Stop(name string) error
	Remove(name string) error
}

func SvcRuntime() (ServiceManager, error) {
	switch runtime.GOOS {
	case "linux":
		return &SystemdManager{}, nil
	case "darwin":
		return &LaunchdManager{}, nil
	case "windows":
		return &NSSMManager{}, nil
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
	default:
		return "systemd on Linux, launchd on macOS, or NSSM on Windows"
	}
}
