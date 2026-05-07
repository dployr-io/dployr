// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dployr-io/dployr/pkg/core/utils"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"

	"github.com/dployr-io/dployr/internal/scripts"
)

type DeployRequest struct{}

// DockerIgnoreContent defines patterns to exclude from Docker builds
const DockerIgnoreContent = `.git/
.gitignore
.gitattributes
.dockerignore
.editorconfig

# AI/IDE tools and caches
.claude/
.cursor/
.aider/
.vscode/
.vscode-server/
.idea/
.anthropic-api-cache/

# Language-specific dependencies (reinstalled in container)
node_modules/
.npm/
.pnpm-store/
package-lock.json
yarn.lock
pnpm-lock.yaml
bun.lockb
vendor/
.venv/
venv/
env/
__pycache__/
*.pyc
.Python
.pytest_cache/
.tox/
.coverage
htmlcov/
dist-info/
Gemfile.lock
.bundle/
.rubygems.lock
/vendor/bundle/

# Build outputs
dist/
build/
out/
bin/
obj/
*.class
*.jar
*.war
.gradle/
.m2/
target/
.cache/
.next/
.nuxt/
out/

# Test coverage and reports
test/
tests/
__tests__/
*.test.js
*.spec.js
*.test.ts
*.spec.ts
coverage/
.nyc_output/
.xunit-results/

# Development environment
.env
.env.local
.env.*.local
.DS_Store
Thumbs.db
*.swp
*.swo
*~
.vagrant/
.vfox.lock

# Documentation and metadata
README.md
CHANGELOG.md
LICENSE
docs/
examples/
CONTRIBUTING.md
.github/
.gitlab-ci.yml
.circleci/
Dockerfile
docker-compose*.yml

# Runtime/temporary
*.log
.git-blame-ignore-revs
.npm-debug.log*
yarn-debug.log*
yarn-error.log*
.eslintcache
.prettierignore
lerna-debug.log*
.pnp.js
.yarn/install-state.gz
`

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
		cloneCmd := fmt.Sprintf("git clone --depth 1 --branch %s %s .", remote.Branch, authUrl)
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

	// Write .dockerignore to exclude build artifacts and dependencies
	if err := writeDockerIgnore(destDir); err != nil {
		return fmt.Errorf("failed to write .dockerignore: %s", err)
	}

	return nil
}

func writeDockerIgnore(destDir string) error {
	dockerIgnorePath := filepath.Join(destDir, ".dockerignore")
	if _, err := os.Stat(dockerIgnorePath); err == nil {
		return nil // Respect existing .dockerignore
	}
	return os.WriteFile(dockerIgnorePath, []byte(DockerIgnoreContent), 0644)
}

// PullImage pulls a docker image from a registry
func PullImage(imageRef string, workDir string, config *shared.Config) error {
	if !isValidDockerImageRef(imageRef) {
		return fmt.Errorf("invalid docker image reference: %s", imageRef)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pullCmd := fmt.Sprintf("docker pull %s", imageRef)
	if err := shared.Exec(ctx, pullCmd, workDir); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("docker pull timed out after 5 minutes")
		}
		return fmt.Errorf("docker pull failed: %s", err)
	}
	return nil
}

// DeployApp handles runtime setup, build, and service installation
func DeployApp(bp store.Blueprint, name, logPath string) error {
	version := string(bp.Runtime.Version)
	if version == "" {
		return fmt.Errorf("runtime version cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	return runDeployScript(ctx, bp, name, logPath)
}

func runDeployScript(ctx context.Context, bp store.Blueprint, name, logPath string) error {
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

	args := []string{
		tmpFile.Name(),
		"deploy",
		bp.Name,
		string(bp.Source),
		string(bp.Type),
		string(bp.Runtime.Type),
		bp.Runtime.Version,
		bp.WorkingDir,
		bp.RunCmd,
		desc,
		buildCmd,
		port,
		strconv.Itoa(utils.ComputeHostPort(bp.Name)),
		bp.Image,
		bp.StaticDir,
	}

	cmd := exec.CommandContext(ctx, "bash", args...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
	)

	// Capture stdout and stderr to deployment log
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	var wg sync.WaitGroup

	// Stream stdout to log file
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			shared.LogInfoF(name, logPath, scanner.Text())
		}
	}()

	// Stream stderr to log file
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			shared.LogWarnF(name, logPath, scanner.Text())
		}
	}()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Wait for streaming to complete
	wg.Wait()

	// Wait for command to finish
	return cmd.Wait()
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

var dockerImageRegex = regexp.MustCompile(`^([a-zA-Z0-9][-a-zA-Z0-9.]*(?::[0-9]+)?/)?([a-zA-Z0-9._/-]+/)?([a-zA-Z0-9._/-]+)(:[a-zA-Z0-9._-]+|@sha256:[a-fA-F0-9]{64})?$`)

func isValidDockerImageRef(ref string) bool {
	return dockerImageRegex.MatchString(ref)
}
