package deploy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"dployr/pkg/core/utils"
	"dployr/pkg/shared"
	"dployr/pkg/store"
)

// SetupDir creates a working directory for the deployment
func SetupDir(name string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not resolve user home directory: %v", err)
	}
	workDir := filepath.Join(homeDir, ".dployr", "services", utils.FormatName(name))
	err = os.MkdirAll(workDir, 0755)
	if err != nil {
		return "", err
	}

	return workDir, nil
}

// CloneRepo clones a git repository to the specified directory
func CloneRepo(remote store.RemoteObj, destDir, workDir string, config *shared.Config) error {
	workDir = fmt.Sprint(destDir, "/", workDir)
	authUrl, err := utils.BuildAuthUrl(remote.Url, config)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if _, err := os.Stat(workDir); err == nil {
		cmd := exec.CommandContext(ctx, "git", "-C", destDir, "pull")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = nil
		if err := cmd.Run(); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("git pull timed out after 5 minutes")
			}
			return err
		}
	} else {
		args := []string{"clone", "--branch", remote.Branch, authUrl, destDir}
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = nil
		if err := cmd.Run(); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("git clone timed out after 5 minutes")
			}
			return fmt.Errorf("git clone failed: %s", err)
		}
	}

	if remote.CommitHash != "" {
		cmd := exec.CommandContext(ctx, "git", "-C", destDir, "checkout", remote.CommitHash)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = nil
		if err := cmd.Run(); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("git checkout timed out after 5 minutes")
			}
			return fmt.Errorf("git checkout failed: %s", err)
		}
	}
	return nil
}

// SetupRuntime sets up the runtime environment using vfox
func SetupRuntime(r store.RuntimeObj, workDir string) error {
	vfox, err := utils.GetVfox()
	if err != nil {
		return fmt.Errorf("failed to find vfox executable: %v", err)
	}

	version := string(r.Version)
	if version == "" {
		return fmt.Errorf("runtime version cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, vfox, "install", string(r.Type)+"@"+version, "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = nil
	cmd.Env = append(os.Environ(), "VFOX_NONINTERACTIVE=1")
	err = cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("vfox command timed out after 5 minutes")
		}
		return fmt.Errorf("vfox command failed: %v", err)
	}

	cmd = exec.CommandContext(ctx, vfox, "use", string(r.Type)+"@"+version)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = nil
	err = cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("vfox command timed out after 5 minutes")
		}
		return fmt.Errorf("vfox command failed: %v", err)
	}

	return nil
}

// InstallDeps installs dependencies using the build command
func InstallDeps(buildCmd, workDir string, r store.RuntimeObj) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	exe, cmdArgs, err := utils.GetExeArgs(r, buildCmd)
	if err != nil {
		return fmt.Errorf("failed to get executable: %v", err)
	}

	cmd := exec.CommandContext(ctx, exe, cmdArgs...)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = nil
	err = cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("build command timed out after 10 minutes")
		}
		return fmt.Errorf("build command failed: %v", err)
	}

	return nil
}
