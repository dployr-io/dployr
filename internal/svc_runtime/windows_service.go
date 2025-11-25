package svc_runtime

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dployr-io/dployr/pkg/core/service"
	"github.com/dployr-io/dployr/pkg/core/utils"
)

type NSSMManager struct{}

// Locate nssm.exe
func (n *NSSMManager) findExe() (string, error) {
	possible := []string{
		`C:\Program Files\nssm\nssm.exe`,
		`C:\Program Files (x86)\nssm\nssm.exe`,
		`C:\ProgramData\chocolatey\bin\nssm.exe`,
		`C:\ProgramData\chocolatey\lib\nssm\tools\nssm.exe`,
		`C:\nssm\nssm.exe`,
		`C:\Tools\nssm\nssm.exe`,
		`C:\nssm.exe`,
	}

	if home, err := os.UserHomeDir(); err == nil {
		possible = append(possible, filepath.Join(home, `scoop\apps\nssm\current\nssm.exe`))
	}

	if path, err := exec.LookPath("nssm.exe"); err == nil {
		possible = append([]string{path}, possible...)
	}

	for _, p := range possible {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("nssm.exe not found — ensure NSSM is installed and in PATH")
}

func (n *NSSMManager) Status(name string) (string, error) {
	nssm, err := n.findExe()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(nssm, "status", name)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get service status: %v - %s", err, out.String())
	}

	switch {
	case strings.Contains(out.String(), "SERVICE_RUNNING"):
		return string(service.SvcRunning), nil
	case strings.Contains(out.String(), "SERVICE_STOPPED"):
		return string(service.SvcStopped), nil
	default:
		return string(service.SvcUnknown), nil
	}
}

func (n *NSSMManager) Install(name, desc, runCmd, workDir string, envVars map[string]string) error {
	nssm, err := n.findExe()
	if err != nil {
		return err
	}

	// Create a batch file for Windows
	batchContent := fmt.Sprintf(`@echo off
%s`, runCmd)

	batFile := filepath.Join(workDir, "service.bat")
	err = os.WriteFile(batFile, []byte(batchContent), 0755)
	if err != nil {
		return fmt.Errorf("failed to create service batch file: %v", err)
	}

	name = utils.FormatName(name)

	// nssm install <name> <batch_file>
	cmd := exec.Command(nssm, "install", name, batFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install service: %v", err)
	}

	exec.Command(nssm, "set", name, "AppDirectory", workDir).Run()
	for k, v := range envVars {
		exec.Command(nssm, "set", name, "AppEnvironmentExtra", fmt.Sprintf("%s=%s", k, v)).Run()
	}
	exec.Command(nssm, "set", name, "Start", "SERVICE_AUTO_START").Run()
	if desc != "" {
		exec.Command(nssm, "set", name, "Description", desc).Run()
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	logDir := filepath.Join(home, ".dployr", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	outLog := filepath.Join(logDir, fmt.Sprintf("%s.log", name))

	exec.Command(nssm, "set", name, "AppStdout", outLog).Run()
	exec.Command(nssm, "set", name, "AppStderr", outLog).Run()
	exec.Command(nssm, "set", name, "AppRotateFiles", "1").Run()
	exec.Command(nssm, "set", name, "AppRotateBytes", "1048576").Run()
	exec.Command(nssm, "set", name, "AppRotateDelay", "86400").Run()

	return nil
}

func (n *NSSMManager) Start(name string) error {
	nssm, err := n.findExe()
	if err != nil {
		return err
	}
	cmd := exec.Command(nssm, "start", utils.FormatName(name))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (n *NSSMManager) Stop(name string) error {
	nssm, err := n.findExe()
	if err != nil {
		return err
	}
	cmd := exec.Command(nssm, "stop", utils.FormatName(name))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (n *NSSMManager) Remove(name string) error {
	nssm, err := n.findExe()
	if err != nil {
		return err
	}
	cmd := exec.Command(nssm, "remove", utils.FormatName(name), "confirm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
