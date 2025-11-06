package svc_runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"dployr/pkg/core/service"
	"dployr/pkg/core/utils"
)

type SystemdManager struct{}

func (s *SystemdManager) Status(name string) (string, error) {
	name = utils.FormatName(name)
	cmd := exec.Command("systemctl", "--user", "is-active", name)
	output, err := cmd.Output()
	if err != nil {
		return string(service.SvcStopped), nil
	}

	status := strings.TrimSpace(string(output))
	switch status {
	case "active":
		return string(service.SvcRunning), nil
	case "inactive", "failed":
		return string(service.SvcStopped), nil
	default:
		return string(service.SvcUnknown), nil
	}
}

func (s *SystemdManager) Install(name, desc, runCmd, workDir string, envVars map[string]string) error {
	name = utils.FormatName(name)

	// Create systemd user directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	systemdDir := filepath.Join(home, ".config", "systemd", "user")
	if err := os.MkdirAll(systemdDir, 0755); err != nil {
		return fmt.Errorf("failed to create systemd directory: %w", err)
	}

	// Create log directory
	logDir := filepath.Join(home, ".dployr", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Build environment variables
	envSection := ""
	for k, v := range envVars {
		envSection += fmt.Sprintf("Environment=%s=%s\n", k, v)
	}

	// Create service file content
	serviceContent := fmt.Sprintf(`[Unit]
Description=%s
After=network.target

[Service]
Type=simple
ExecStart=%s
WorkingDirectory=%s
%sRestart=always
RestartSec=10
StandardOutput=append:%s/%s.log
StandardError=append:%s/%s.log

[Install]
WantedBy=default.target
`, desc, runCmd, workDir, envSection, logDir, name, logDir, name)

	// Write service file
	serviceFile := filepath.Join(systemdDir, name+".service")
	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Reload systemd and enable service
	if err := exec.Command("systemctl", "--user", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	if err := exec.Command("systemctl", "--user", "enable", name).Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	return nil
}

func (s *SystemdManager) Start(name string) error {
	name = utils.FormatName(name)
	cmd := exec.Command("systemctl", "--user", "start", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *SystemdManager) Stop(name string) error {
	name = utils.FormatName(name)
	cmd := exec.Command("systemctl", "--user", "stop", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *SystemdManager) Remove(name string) error {
	name = utils.FormatName(name)

	// Stop and disable service
	exec.Command("systemctl", "--user", "stop", name).Run()
	exec.Command("systemctl", "--user", "disable", name).Run()

	// Remove service file
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	serviceFile := filepath.Join(home, ".config", "systemd", "user", name+".service")
	if err := os.Remove(serviceFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	// Reload systemd
	return exec.Command("systemctl", "--user", "daemon-reload").Run()
}
