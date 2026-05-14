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

type DockerService struct{}

func runScript(scriptContent string, args ...string) error {
	tmpFile, err := os.CreateTemp("", "dployr*.sh")
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

func (d *DockerService) Status(name string) (string, error) {
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

func (d *DockerService) Install(name, desc, runCmd, workDir string, envVars map[string]string) error {
	// Docker installations are handled by the deployment script (deploy_app.sh or docker.sh).
	// This method is not used in the current workflow.
	return nil
}

func (d *DockerService) Start(name string) error {
	return runScript(scripts.DockerScript, "start", name)
}

func (d *DockerService) Stop(name string) error {
	return runScript(scripts.DockerScript, "stop", name)
}

func (d *DockerService) Remove(name string) error {
	return runScript(scripts.DockerScript, "remove", name)
}

func (d *DockerService) HealthStatus(name string) (string, error) {
	out, err := exec.Command("docker", "inspect", "--format", "{{.State.Health.Status}}", name).Output()
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(out)), nil
}

// Ice stops the container and removes its image to free up disk space.
// The container configuration is preserved in the service store so it can be redeployed.
func (d *DockerService) Ice(name string) error {
	// Capture image name before stopping so we can remove it afterward.
	imageOut, _ := exec.Command("docker", "inspect", "--format", "{{.Config.Image}}", name).Output()
	image := strings.TrimSpace(string(imageOut))

	if err := exec.Command("docker", "stop", name).Run(); err != nil {
		return fmt.Errorf("failed to stop container %s: %w", name, err)
	}

	if image != "" {
		// Best-effort: ignore errors (image may be shared or already removed).
		exec.Command("docker", "rmi", image).Run()
	}

	return nil
}
