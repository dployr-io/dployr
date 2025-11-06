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

type LaunchdManager struct{}

func (l *LaunchdManager) Status(name string) (string, error) {
	name = utils.FormatName(name)
	label := fmt.Sprintf("user.%s", name)

	cmd := exec.Command("launchctl", "list", label)
	output, err := cmd.Output()
	if err != nil {
		return string(service.SvcStopped), nil
	}

	if strings.Contains(string(output), label) {
		return string(service.SvcRunning), nil
	}
	return string(service.SvcStopped), nil
}

func (l *LaunchdManager) Install(name, desc, runCmd, workDir string, envVars map[string]string) error {
	name = utils.FormatName(name)
	label := fmt.Sprintf("user.%s", name)

	// Create LaunchAgents directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	launchDir := filepath.Join(home, "Library", "LaunchAgents")
	if err := os.MkdirAll(launchDir, 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	// Create log directory
	logDir := filepath.Join(home, ".dployr", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Parse command and arguments
	parts := strings.Fields(runCmd)
	if len(parts) == 0 {
		return fmt.Errorf("runCmd cannot be empty")
	}

	program := parts[0]
	args := parts[1:]

	// Build environment variables
	envDict := ""
	for k, v := range envVars {
		envDict += fmt.Sprintf(`		<key>%s</key>
		<string>%s</string>
`, k, v)
	}

	// Build program arguments
	argsArray := fmt.Sprintf(`		<string>%s</string>`, program)
	for _, arg := range args {
		argsArray += fmt.Sprintf(`
		<string>%s</string>`, arg)
	}

	// Create plist content
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
%s
	</array>
	<key>WorkingDirectory</key>
	<string>%s</string>
	<key>EnvironmentVariables</key>
	<dict>
%s	</dict>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>%s/%s.log</string>
	<key>StandardErrorPath</key>
	<string>%s/%s.log</string>
</dict>
</plist>
`, label, argsArray, workDir, envDict, logDir, name, logDir, name)

	// Write plist file
	plistFile := filepath.Join(launchDir, label+".plist")
	if err := os.WriteFile(plistFile, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	// Load the service
	if err := exec.Command("launchctl", "load", plistFile).Run(); err != nil {
		return fmt.Errorf("failed to load service: %w", err)
	}

	return nil
}

func (l *LaunchdManager) Start(name string) error {
	name = utils.FormatName(name)
	label := fmt.Sprintf("user.%s", name)

	cmd := exec.Command("launchctl", "start", label)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (l *LaunchdManager) Stop(name string) error {
	name = utils.FormatName(name)
	label := fmt.Sprintf("user.%s", name)

	cmd := exec.Command("launchctl", "stop", label)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (l *LaunchdManager) Remove(name string) error {
	name = utils.FormatName(name)
	label := fmt.Sprintf("user.%s", name)

	// Stop and unload service
	exec.Command("launchctl", "stop", label).Run()
	exec.Command("launchctl", "unload", label).Run()

	// Remove plist file
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	plistFile := filepath.Join(home, "Library", "LaunchAgents", label+".plist")
	if err := os.Remove(plistFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plist file: %w", err)
	}

	return nil
}
