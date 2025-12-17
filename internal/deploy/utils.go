// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dployr-io/dployr/pkg/core/utils"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"

	"github.com/dployr-io/dployr/internal/scripts"
)

type DeployRequest struct{}

// SetupDir creates a working directory for the deployment
func SetupDir(name string) (string, error) {
	dataDir := utils.GetDataDir()
	workDir := filepath.Join(dataDir, ".dployr", "services", utils.FormatName(name))
	err := os.MkdirAll(workDir, 0755)
	if err != nil {
		return "", err
	}

	return workDir, nil
}

// CloneRepo clones a git repository to the specified directory
func CloneRepo(remote store.RemoteObj, destDir, workDir string, config *shared.Config) error {
	authUrl, err := buildAuthUrl(remote.Url)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Check if destDir already has a .git directory
	gitDir := filepath.Join(destDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		pullCmd := fmt.Sprintf("git -C %s pull", destDir)
		if err := shared.Exec(ctx, pullCmd, destDir); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("git pull timed out after 5 minutes")
			}
			return err
		}
	} else {
		// Ensure destDir exists
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create destination directory: %s", err)
		}
		cloneCmd := fmt.Sprintf("git clone --branch %s %s .", remote.Branch, authUrl)
		if err := shared.Exec(ctx, cloneCmd, destDir); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("git clone timed out after 5 minutes")
			}
			return fmt.Errorf("git clone failed: %s", err)
		}
	}

	if remote.CommitHash != "" {
		checkoutCmd := fmt.Sprintf("git -C %s checkout %s", destDir, remote.CommitHash)
		if err := shared.Exec(ctx, checkoutCmd, "."); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("git checkout timed out after 5 minutes")
			}
			return fmt.Errorf("git checkout failed: %s", err)
		}
	}
	return nil
}

// DeployApp handles runtime setup, build, and service installation
func DeployApp(bp store.Blueprint) error {
	version := string(bp.Runtime.Version)
	if version == "" {
		return fmt.Errorf("runtime version cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	return runDeployScript(ctx, bp)
}

func runDeployScript(ctx context.Context, bp store.Blueprint) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("unified deployment script not yet supported on Windows")
	}

	// Use service name as description if empty
	desc := bp.Desc
	if desc == "" {
		desc = fmt.Sprintf("%s service", bp.Name)
	}

	// Write service config.toml with env vars and secrets
	if err := writeServiceConfig(bp); err != nil {
		return fmt.Errorf("failed to write service config: %v", err)
	}

	// Create temporary script file
	tmpFile, err := os.CreateTemp("", "deploy_app*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temp script: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write script content
	if _, err := tmpFile.WriteString(scripts.DeployScript); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write script: %v", err)
	}
	tmpFile.Close()

	// Make script executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %v", err)
	}

	// Build arguments - always include all positional args
	buildCmd := bp.BuildCmd
	if buildCmd == "" {
		buildCmd = ""
	}

	port := fmt.Sprintf("%d", bp.Port)
	if bp.Port == 0 {
		port = "3000"
	}

	args := []string{tmpFile.Name(), "deploy", bp.Name, string(bp.Runtime.Type), bp.Runtime.Version, bp.WorkingDir, bp.RunCmd, desc, buildCmd, port}

	cmd := exec.CommandContext(ctx, "bash", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
	)

	return cmd.Run()
}

// ServiceConfig represents the TOML structure for service environment configuration
type ServiceConfig struct {
	Env     map[string]string `toml:"env"`
	Secrets map[string]string `toml:"secrets"`
}

// writeServiceConfig writes the service config.toml file with env vars and secrets
func writeServiceConfig(bp store.Blueprint) error {
	configDir := filepath.Join(bp.WorkingDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	configPath := filepath.Join(configDir, "config.toml")

	config := ServiceConfig{
		Env:     bp.EnvVars,
		Secrets: bp.Secrets,
	}

	// Initialize empty maps if nil to ensure proper TOML output
	if config.Env == nil {
		config.Env = make(map[string]string)
	}
	if config.Secrets == nil {
		config.Secrets = make(map[string]string)
	}

	// Build TOML content manually to maintain control over format
	var content strings.Builder
	content.WriteString("[env]\n")
	for key, value := range config.Env {
		content.WriteString(fmt.Sprintf("%s = %q\n", key, value))
	}
	content.WriteString("\n[secrets]\n")
	for key, value := range config.Secrets {
		content.WriteString(fmt.Sprintf("%s = %q\n", key, value))
	}

	if err := os.WriteFile(configPath, []byte(content.String()), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

func buildAuthUrl(url string) (string, error) {
	if strings.Contains(url, "@") {
		return url, nil
	}
	var token, username string

	switch {
	}

	if token == "" {
		return url, nil
	}

	cleanUrl := url
	if after, ok := strings.CutPrefix(cleanUrl, "http://"); ok {
		cleanUrl = "https://" + after
	}
	if strings.HasPrefix(cleanUrl, "https://") {
		return strings.Replace(cleanUrl, "https://", fmt.Sprintf("https://%s:%s@", username, token), 1), nil
	}
	return url, nil
}
