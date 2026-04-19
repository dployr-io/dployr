// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package svc_runtime

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dployr-io/dployr/internal/scripts"
)

type DockerManager struct{}

func (d *DockerManager) runScript(scriptContent string, args ...string) error {
	tmpFile, err := os.CreateTemp("", "docker*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temp script: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(scriptContent); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write script: %v", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %v", err)
	}

	cmd := exec.Command("bash", append([]string{tmpFile.Name()}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (d *DockerManager) Status(name string) (string, error) {
	tmpFile, err := os.CreateTemp("", "docker*.sh")
	if err != nil {
		return "", fmt.Errorf("failed to create temp script: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(scripts.DockerScript); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write script: %v", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return "", fmt.Errorf("failed to make script executable: %v", err)
	}

	cmd := exec.Command("bash", tmpFile.Name(), "status", name)
	output, err := cmd.Output()

	status := strings.TrimSpace(string(output))
	if err != nil {
		if status == "stopped" {
			return "", fmt.Errorf("service %s does not exist", name)
		}
		return "", fmt.Errorf("failed to check service status: %v", err)
	}
	return status, nil
}

func (d *DockerManager) Install(name, desc, runCmd, workDir string, envVars map[string]string) error {
	// Docker installations are handled by the deployment script (deploy_app.sh or docker.sh).
	// This method is not used in the current workflow.
	return nil
}

func (d *DockerManager) Start(name string) error {
	return d.runScript(scripts.DockerScript, "start", name)
}

func (d *DockerManager) Stop(name string) error {
	return d.runScript(scripts.DockerScript, "stop", name)
}

func (d *DockerManager) Remove(name string) error {
	return d.runScript(scripts.DockerScript, "remove", name)
}
