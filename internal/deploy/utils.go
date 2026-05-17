// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
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

func imageRef(registryURL, name string) string {
	slug := strings.ToLower(strings.ReplaceAll(name, "_", "-"))
	tag := fmt.Sprintf("%s-%d", slug, time.Now().UnixMilli())
	return fmt.Sprintf("%s/apps:%s", strings.TrimRight(registryURL, "/"), tag)
}

type BuildOpts struct {
	Runtime         string
	Version         string
	BuildCmd        string
	RunCmd          string
	Port            int
	IsNextJS        bool
	HealthCheckPath string // optional; defaults to root URL check if empty
}

// detectNextJS returns true if the directory looks like a Next.js project —
// either a next.config.* file exists or package.json lists "next" as a dependency.
func detectNextJS(dir string) bool {
	patterns := []string{"next.config.js", "next.config.ts", "next.config.mjs", "next.config.cjs"}
	for _, p := range patterns {
		if _, err := os.Stat(filepath.Join(dir, p)); err == nil {
			return true
		}
	}
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return false
	}
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}
	_, inDeps := pkg.Dependencies["next"]
	_, inDev := pkg.DevDependencies["next"]
	return inDeps || inDev
}

func BuildImage(name, srcDir string, cfg *shared.Config, opts BuildOpts) (string, error) {
	if cfg.RegistryURL == "" {
		return "", fmt.Errorf("REGISTRY_URL is not configured on this build node")
	}

	ref := imageRef(cfg.RegistryURL, name)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	if cfg.RegistryAuth != "" {
		registry := strings.SplitN(ref, "/", 2)[0]
		if err := registryLogin(ctx, registry, cfg.RegistryAuth, srcDir); err != nil {
			return "", fmt.Errorf("registry login failed: %w", err)
		}
	}

	if err := ensureDockerfile(srcDir, opts); err != nil {
		return "", fmt.Errorf("dockerfile setup failed: %w", err)
	}

	buildCmd := fmt.Sprintf("docker build --tag %s .", ref)
	if err := shared.Exec(ctx, buildCmd, srcDir); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("docker build timed out after 20 minutes")
		}
		return "", fmt.Errorf("docker build failed: %w", err)
	}

	pushCmd := fmt.Sprintf("docker push %s", ref)
	if err := shared.Exec(ctx, pushCmd, srcDir); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("docker push timed out")
		}
		return "", fmt.Errorf("docker push failed: %w", err)
	}

	_ = shared.Exec(ctx, fmt.Sprintf("docker rmi %s", ref), srcDir)

	return ref, nil
}

// ensureDockerfile writes a generated Dockerfile unless the repo already ships one.
// It checks git to distinguish committed Dockerfiles from stale generated ones left
// by a previous build attempt — stale files are always overwritten.
func ensureDockerfile(dir string, opts BuildOpts) error {
	dockerfilePath := filepath.Join(dir, "Dockerfile")

	// Only skip generation if the Dockerfile is tracked by git (committed by the user).
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "ls-files", "--error-unmatch", "Dockerfile")
	if err := cmd.Run(); err == nil {
		return nil // committed Dockerfile — respect it
	}

	content := generateDockerfile(opts)
	return os.WriteFile(dockerfilePath, []byte(content), 0644)
}

func runtimeBaseImage(runtime, version string) string {
	switch runtime {
	case "golang":
		return "golang:" + version
	case "php":
		return "php:" + version + "-fpm-alpine"
	case "python":
		return "python:" + version + "-slim"
	case "nodejs":
		return "node:" + version
	case "ruby":
		return "ruby:" + version + "-slim"
	case "dotnet":
		return "mcr.microsoft.com/dotnet:" + version
	case "java":
		return "maven:3-eclipse-temurin-" + version
	default:
		return runtime + ":" + version
	}
}

// healthCheck returns the HEALTHCHECK instruction for the given opts.
// If HealthCheckPath is set, it issues a strict spider check (expects 2xx).
// Otherwise it uses the root URL but accepts any HTTP response (including 404)
// as "alive" — only a connection failure marks the container unhealthy.
func healthCheck(opts BuildOpts) string {
	const prefix = "HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 CMD "
	path := opts.HealthCheckPath
	if path == "" {
		// Lenient: grep for any HTTP response line so 404 is still treated as alive.
		return prefix + `wget -qSO /dev/null http://localhost:${PORT}/ 2>&1 | grep -qE "^\s+HTTP/"` + "\n"
	}
	return prefix + fmt.Sprintf("wget --spider -q http://localhost:${PORT}%s || exit 1", path) + "\n"
}

func generateDockerfile(opts BuildOpts) string {
	image := runtimeBaseImage(opts.runtime(), opts.version())
	port := opts.Port
	if port == 0 {
		port = 8080
	}

	var b strings.Builder
	fmt.Fprintf(&b, "FROM %s\n\nWORKDIR /app\n\n", image)

	// templateInstall is the install command already baked into each runtime's
	// template layer. We skip BuildCmd if it would just re-run the same thing.
	var templateInstall string
	alpineTag := func(ver string) string {
		if ver == "" || ver == "lts" {
			return "node:lts-alpine"
		}
		return "node:" + ver + "-alpine"
	}

	switch opts.runtime() {
	case "nodejs":
		ver := opts.version()
		if opts.IsNextJS {
			fmt.Fprintf(&b, "COPY package*.json ./\nRUN npm install\nCOPY . .\n")
			buildCmd := opts.BuildCmd
			if buildCmd == "" {
				buildCmd = "npm run build"
			}
			fmt.Fprintf(&b, "RUN %s\n", buildCmd)
			fmt.Fprintf(&b, "\nFROM %s AS runner\nWORKDIR /app\n", alpineTag(ver))
			b.WriteString("COPY --from=0 /app/.next/standalone ./\n")
			b.WriteString("COPY --from=0 /app/.next/static ./.next/static\n")
			b.WriteString("COPY --from=0 /app/public ./public\n")
			fmt.Fprintf(&b, "\nENV PORT=%d\n%sCMD [\"node\", \"server.js\"]\n", port, healthCheck(opts))
			return b.String()
		}
		if opts.BuildCmd != "" {
			// Multi-stage: full image to build, alpine to run.
			b.WriteString("COPY package*.json ./\nRUN npm install\nCOPY . .\n")
			fmt.Fprintf(&b, "RUN %s\n", opts.BuildCmd)
			fmt.Fprintf(&b, "\nFROM %s AS runner\nWORKDIR /app\n", alpineTag(ver))
			b.WriteString("COPY package*.json ./\nRUN npm install --omit=dev\n")
			b.WriteString("COPY --from=0 /app ./\n")
			fmt.Fprintf(&b, "\nENV PORT=%d\n%s", port, healthCheck(opts))
			if opts.RunCmd != "" {
				fmt.Fprintf(&b, "CMD %s\n", opts.RunCmd)
			}
			return b.String()
		}
		// No build step — install prod deps only on alpine directly.
		b.Reset()
		fmt.Fprintf(&b, "FROM %s\n\nWORKDIR /app\n\n", alpineTag(ver))
		b.WriteString("COPY package*.json ./\nRUN npm install --omit=dev\nCOPY . .\n")
		templateInstall = "npm install --omit=dev"
	case "python":
		b.WriteString("COPY requirements.txt ./\nRUN pip install --no-cache-dir -r requirements.txt\nCOPY . .\n")
		templateInstall = "pip install --no-cache-dir -r requirements.txt"
	case "golang":
		b.WriteString("COPY go.mod go.sum ./\nRUN go mod download\nCOPY . .\n")
		buildTarget := "."
		if opts.RunCmd != "" {
			buildTarget = opts.RunCmd
		}
		fmt.Fprintf(&b, "RUN CGO_ENABLED=0 go build -o /bin/app %s\n", buildTarget)
		b.WriteString("\nFROM alpine:3\n")
		b.WriteString("RUN apk --no-cache add ca-certificates tzdata\n")
		b.WriteString("COPY --from=0 /bin/app /bin/app\n")
		fmt.Fprintf(&b, "\nENV PORT=%d\n%sCMD [\"/bin/app\"]\n", port, healthCheck(opts))
		return b.String()
	case "ruby":
		b.WriteString("COPY Gemfile* ./\nRUN bundle config set --local without 'development test' && bundle install\nCOPY . .\n")
		templateInstall = "bundle install"
	case "java":
		buildCmd := opts.BuildCmd
		if buildCmd == "" {
			buildCmd = "mvn package -DskipTests"
		}
		b.WriteString("COPY pom.xml ./\nRUN mvn dependency:go-offline -B\nCOPY . .\n")
		fmt.Fprintf(&b, "RUN %s\n", buildCmd)
		fmt.Fprintf(&b, "\nFROM eclipse-temurin:%s-jre-alpine\nWORKDIR /app\n", opts.version())
		b.WriteString("COPY --from=0 /app/target/*.jar app.jar\n")
		fmt.Fprintf(&b, "\nENV PORT=%d\n%sCMD [\"java\", \"-jar\", \"app.jar\"]\n", port, healthCheck(opts))
		return b.String()
	default:
		b.WriteString("COPY . .\n")
	}

	if opts.BuildCmd != "" && strings.TrimSpace(opts.BuildCmd) != templateInstall {
		fmt.Fprintf(&b, "RUN %s\n", opts.BuildCmd)
	}
	fmt.Fprintf(&b, "\nENV PORT=%d\n%s", port, healthCheck(opts))
	if opts.RunCmd != "" {
		fmt.Fprintf(&b, "CMD %s\n", opts.RunCmd)
	}
	return b.String()
}

func (o BuildOpts) runtime() string {
	if o.Runtime == "" {
		return "nodejs"
	}
	return o.Runtime
}

func (o BuildOpts) version() string {
	if o.Version == "" {
		return "lts"
	}
	return o.Version
}

// registryLogin authenticates Docker against the given registry.
// authB64 must be a base64-encoded JSON blob: {"username":"…","password":"…"}.
// Falls back to treating the value as a raw password with username "token" for
// registries (like DigitalOcean) that accept any username.
func registryLogin(ctx context.Context, registry, authB64, workDir string) error {
	username, password := "token", authB64

	raw, err := base64.StdEncoding.DecodeString(authB64)
	if err == nil {
		decoded := string(raw)
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if json.Unmarshal(raw, &creds) == nil && creds.Password != "" {
			username, password = creds.Username, creds.Password
		} else if u, p, ok := strings.Cut(decoded, ":"); ok {
			// base64("username:password") — standard Docker Basic Auth format
			username, password = u, p
		} else {
			// bare token — use as password
			password = decoded
		}
	}

	cmd := exec.CommandContext(ctx, "docker", "login", "--username", username, "--password-stdin", registry)
	cmd.Dir = workDir
	cmd.Stdin = strings.NewReader(password)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("exit status %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

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
func CloneRepo(remote store.RemoteObj, destDir string, config *shared.Config) error {
	authUrl, err := buildAuthUrl(remote.Url, remote.Token)
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

	if config != nil && config.RegistryAuth != "" {
		registry := strings.SplitN(imageRef, "/", 2)[0]
		if err := registryLogin(ctx, registry, config.RegistryAuth, workDir); err != nil {
			return fmt.Errorf("registry login failed: %w", err)
		}
	}

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
func DeployApp(bp store.Blueprint, name, logPath string, cfg *shared.Config) error {
	version := string(bp.Runtime.Version)
	if version == "" {
		return fmt.Errorf("runtime version cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	return runDeployScript(ctx, bp, name, logPath, cfg)
}

func runDeployScript(ctx context.Context, bp store.Blueprint, name, logPath string, cfg *shared.Config) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("unified deployment script not yet supported on Windows")
	}

	desc := bp.Desc
	if desc == "" {
		desc = fmt.Sprintf("%s service", bp.Name)
	}

	if err := writeServiceConfig(bp); err != nil {
		return fmt.Errorf("failed to write service config: %v", err)
	}

	tmpFile, err := os.CreateTemp("", "deploy_app*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temp script: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(scripts.DeployScript); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write script: %v", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %v", err)
	}

	buildCmd := bp.BuildCmd

	port := fmt.Sprintf("%d", bp.Port)
	if bp.Port == 0 {
		port = "3000"
	}

	memory, cpu, storage, buildMemory := 0, 0, 0, 0
	if cfg != nil {
		memory = cfg.ContainerMemory
		cpu = cfg.ContainerCPU
		storage = cfg.ContainerStorage
		buildMemory = cfg.BuildMemory
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
		strconv.Itoa(memory),
		strconv.Itoa(cpu),
		strconv.Itoa(storage),
		strconv.Itoa(buildMemory),
	}

	cmd := exec.CommandContext(ctx, "bash", args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("HOME=%s", os.Getenv("HOME")))

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
	var scriptErr string // last [ERROR] line written by abort()

	// Stream stdout to log file
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			shared.LogInfoF(name, logPath, scanner.Text())
		}
	}()

	// Stream stderr to log file; surface [ERROR] lines for structured logging
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			shared.LogWarnF(name, logPath, line)
			if strings.HasPrefix(line, "[ERROR]") {
				scriptErr = strings.TrimPrefix(line, "[ERROR] ")
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		if scriptErr != "" {
			return fmt.Errorf("%w: %s", err, scriptErr)
		}
		return err
	}
	return nil
}

func writeServiceConfig(bp store.Blueprint) error {
	if err := os.MkdirAll(bp.WorkingDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	env := bp.EnvVars
	if env == nil {
		env = make(map[string]string)
	}
	secrets := bp.Secrets
	if secrets == nil {
		secrets = make(map[string]string)
	}

	var b strings.Builder

	b.WriteString("[env]\n")
	for k, v := range env {
		fmt.Fprintf(&b, "%s = %q\n", k, v)
	}

	b.WriteString("\n[secrets]\n")
	for k, v := range secrets {
		fmt.Fprintf(&b, "%s = %q\n", k, v)
	}

	if bp.HealthCheck != nil && bp.HealthCheck.Path != "" && bp.Type != store.TypeStatic {
		hc := bp.HealthCheck
		interval, timeout, retries := hc.Interval, hc.Timeout, hc.Retries
		if interval <= 0 {
			interval = 30
		}
		if timeout <= 0 {
			timeout = 5
		}
		if retries <= 0 {
			retries = 3
		}
		fmt.Fprintf(&b, "\n[health_check]\npath = %q\ninterval = %d\ntimeout = %d\nretries = %d\n",
			hc.Path, interval, timeout, retries)
	}

	return os.WriteFile(filepath.Join(bp.WorkingDir, "config.toml"), []byte(b.String()), 0600)
}

func buildAuthUrl(url, token string) (string, error) {
	if strings.Contains(url, "@") {
		return url, nil // credentials already embedded
	}

	if token == "" {
		return url, nil
	}

	// Normalise to HTTPS — git over SSH cannot embed credentials in the URL.
	cleanUrl := url
	if after, ok := strings.CutPrefix(cleanUrl, "http://"); ok {
		cleanUrl = "https://" + after
	}

	if !strings.HasPrefix(cleanUrl, "https://") {
		return url, nil
	}

	var username string
	switch {
	case strings.Contains(cleanUrl, "github.com"):
		username = "x-access-token"
	case strings.Contains(cleanUrl, "gitlab.com"):
		username = "oauth2"
	case strings.Contains(cleanUrl, "bitbucket.org"):
		username = "x-token-auth"
	default:
		username = "oauth2" // reasonable default for self-hosted providers
	}

	return strings.Replace(cleanUrl, "https://", fmt.Sprintf("https://%s:%s@", username, token), 1), nil
}

var dockerImageRegex = regexp.MustCompile(`^([a-zA-Z0-9][-a-zA-Z0-9.]*(?::[0-9]+)?/)?([a-zA-Z0-9._/-]+/)?([a-zA-Z0-9._/-]+)(:[a-zA-Z0-9._-]+|@sha256:[a-fA-F0-9]{64})?$`)

func isValidDockerImageRef(ref string) bool {
	return dockerImageRegex.MatchString(ref)
}
