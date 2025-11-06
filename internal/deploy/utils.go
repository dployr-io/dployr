package deploy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
func SetupRuntime(r store.RuntimeObj, workDir string) error {

	version := string(r.Version)
	if version == "" {
		return fmt.Errorf("runtime version cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Install runtime version
	cmd := fmt.Sprintf("vfox install %s@%s -y", string(r.Type), version)
	err := shared.Exec(ctx, cmd, utils.GetDataDir())
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("vfox command timed out after 5 minutes")
		}
		return fmt.Errorf("vfox command failed: %v", err)
	}

	useCmd := fmt.Sprintf("vfox use %s@%s", string(r.Type), version)
	err = shared.Exec(ctx, useCmd, utils.GetDataDir())
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("vfox command timed out after 5 minutes")
		}
		return fmt.Errorf("vfox command failed: %v", err)
	}

	fmt.Printf("Runtime %s@%s installed and verified successfully\n", r.Type, version)
	return nil
}

// InstallDeps installs dependencies using the build command
func InstallDeps(buildCmd, workDir string, r store.RuntimeObj) error {
	if buildCmd == "" {
		fmt.Printf("No build command specified, skipping dependency installation\n")
		return nil
	}

	fmt.Printf("Installing dependencies with command: %s\n", buildCmd)
	fmt.Printf("Runtime: %s@%s, Working directory: %s\n", r.Type, r.Version, workDir)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	err := shared.Exec(ctx, buildCmd, workDir)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("build command timed out after 10 minutes")
		}
		return fmt.Errorf("build command '%s' failed: %v", buildCmd, err)
	}

	fmt.Printf("Dependencies installed successfully\n")
	return nil
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
