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

	"dployr/internal/scripts"
	"dployr/pkg/core/utils"
	"dployr/pkg/shared"
	"dployr/pkg/store"
)

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
	workDir = fmt.Sprint(destDir, "/", workDir)
	authUrl, err := buildAuthUrl(remote.Url, config)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if _, err := os.Stat(workDir); err == nil {
		pullCmd := fmt.Sprintf("git -C %s pull", destDir)
		if err := shared.Exec(ctx, pullCmd, "."); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("git pull timed out after 5 minutes")
			}
			return err
		}
	} else {
		cloneCmd := fmt.Sprintf("git clone --branch %s %s %s", remote.Branch, authUrl, destDir)
		if err := shared.Exec(ctx, cloneCmd, "."); err != nil {
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

// SetupRuntime sets up the runtime environment using vfox
func SetupRuntime(r store.RuntimeObj, workDir, buildCmd string) error {
	version := string(r.Version)
	if version == "" {
		return fmt.Errorf("runtime version cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	return runRuntimeScript(ctx, string(r.Type), version, workDir, buildCmd)
}

func runRuntimeScript(ctx context.Context, runtimeName, version, workDir, buildCmd string) error {
	var shell string
	var scriptContent string
	var scriptExt string

	if runtime.GOOS == "windows" {
		shell = "pwsh"
		scriptContent = scripts.PowershellScript
		scriptExt = ".ps1"
	} else {
		shell = "bash"
		scriptContent = scripts.BashScript
		scriptExt = ".sh"
	}

	// Create temporary script file
	tmpFile, err := os.CreateTemp("", "setup_runtime*"+scriptExt)
	if err != nil {
		return fmt.Errorf("failed to create temp script: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write script content
	if _, err := tmpFile.WriteString(scriptContent); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write script: %v", err)
	}
	tmpFile.Close()

	// Make script executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
			return fmt.Errorf("failed to make script executable: %v", err)
		}
	}

	var args []string
	if runtime.GOOS == "windows" {
		args = []string{"-File", tmpFile.Name(), runtimeName, version, workDir, buildCmd}
	} else {
		args = []string{tmpFile.Name(), runtimeName, version, workDir, buildCmd}
	}

	cmd := exec.CommandContext(ctx, shell, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("VFOX_HOME=%s/.version-fox", os.Getenv("HOME")),
	)

	return cmd.Run()
}

func buildAuthUrl(url string, config *shared.Config) (string, error) {
	if strings.Contains(url, "@") {
		return url, nil
	}
	var token, username string

	switch {
	case strings.Contains(url, "github.com"):
		token, username = config.GitHubToken, "git"
	case strings.Contains(url, "gitlab.com"):
		token, username = config.GitLabToken, "git"
	case strings.Contains(url, "bitbucket.org"):
		token, username = config.BitBucketToken, "git"
	default:
		return url, nil
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
