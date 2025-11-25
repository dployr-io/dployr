package svc_runtime

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dployr-io/dployr/pkg/core/utils"

	"github.com/dployr-io/dployr/internal/scripts"
)

type SystemdManager struct{}

func (s *SystemdManager) runScript(scriptContent string, args ...string) error {
	tmpFile, err := os.CreateTemp("", "systemd*.sh")
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

func (s *SystemdManager) Status(name string) (string, error) {
	name = utils.FormatName(name)

	tmpFile, err := os.CreateTemp("", "systemd*.sh")
	if err != nil {
		return "", fmt.Errorf("failed to create temp script: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(scripts.SystemdScript); err != nil {
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

func (s *SystemdManager) Install(name, desc, runCmd, workDir string, envVars map[string]string) error {
	name = utils.FormatName(name)
	return s.runScript(scripts.SystemdScript, "install", name, desc, runCmd, workDir)
}

func (s *SystemdManager) Start(name string) error {
	name = utils.FormatName(name)
	return s.runScript(scripts.SystemdScript, "start", name)
}

func (s *SystemdManager) Stop(name string) error {
	name = utils.FormatName(name)
	return s.runScript(scripts.SystemdScript, "stop", name)
}

func (s *SystemdManager) Remove(name string) error {
	name = utils.FormatName(name)
	return s.runScript(scripts.SystemdScript, "remove", name)
}
