// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/pkg/archive"
	specs "github.com/opencontainers/image-spec/specs-go/v1"

	coreutils "github.com/dployr-io/dployr/pkg/core/utils"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"

	"github.com/dployr-io/dployr/internal/scripts"
)

// deployDockerAPI is the subset of the Docker client used by this package.
type deployDockerAPI interface {
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *specs.Platform, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)
	ImageBuild(ctx context.Context, buildContext io.Reader, options dockertypes.ImageBuildOptions) (dockertypes.ImageBuildResponse, error)
	ImagePush(ctx context.Context, image string, options image.PushOptions) (io.ReadCloser, error)
	ImageRemove(ctx context.Context, imageID string, options image.RemoveOptions) ([]image.DeleteResponse, error)
}

// imageRef constructs a uniqe docker image ref
func imageRef(registryURL, name string) string {
	slug := strings.ToLower(strings.ReplaceAll(name, "_", "-"))
	tag := fmt.Sprintf("%s-%d", slug, time.Now().UnixMilli())
	return fmt.Sprintf("%s/apps:%s", strings.TrimRight(registryURL, "/"), tag)
}

type BuildOpts struct {
	Runtime      string
	Version      string
	BuilderImage string // resolved primary FROM image; falls back to runtimeBaseImage when empty
	RunnerImage  string // resolved runner FROM image for multi-stage builds; falls back when empty
	BuildCmd     string
	RunCmd       string
	Port         int
	IsNextJS     bool
	Env          map[string]string
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

// normalizeDockerLine cleans a raw Docker build stream line for display.
// Docker sends multi-line Dockerfile commands (Step N/M : RUN ...) as a single
// stream message with embedded backslash-newline continuations. Replacing them
// with a space prevents display artifacts like "nodejsnpm" or "&&npm".
func normalizeDockerLine(s string) string {
	s = strings.ReplaceAll(s, "\\\n", " ")
	return strings.TrimSpace(s)
}

// writeLogEntry writes a single structured JSON log line to f.
// Uses json.Marshal so string values are always valid JSON regardless of
// their content (e.g. ANSI escape sequences, backslashes).
func writeLogEntry(f *os.File, level, msg string) {
	type entry struct {
		Time  string `json:"time"`
		Level string `json:"level"`
		Msg   string `json:"msg"`
	}
	b, err := json.Marshal(entry{
		Time:  time.Now().UTC().Format(time.RFC3339Nano),
		Level: level,
		Msg:   msg,
	})
	if err != nil {
		return
	}
	b = append(b, '\n')
	f.Write(b) //nolint:errcheck
}

func BuildImage(name, srcDir string, cfg *shared.Config, opts BuildOpts, dockerCli deployDockerAPI, svcName, logDir string) (string, error) {
	if cfg.RegistryURL == "" {
		return "", fmt.Errorf("REGISTRY_URL is not configured on this build node")
	}

	ref := imageRef(cfg.RegistryURL, name)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	if err := ensureDockerfile(srcDir, opts); err != nil {
		return "", fmt.Errorf("dockerfile setup failed: %w", err)
	}

	// Build tar context from srcDir
	var excludes []string
	for line := range strings.SplitSeq(DockerIgnoreContent, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || line == "Dockerfile" || line == ".dockerignore" {
			continue
		}
		excludes = append(excludes, line)
	}
	buildCtx, err := archive.TarWithOptions(srcDir, &archive.TarOptions{ExcludePatterns: excludes})
	if err != nil {
		return "", fmt.Errorf("failed to create build context: %w", err)
	}

	buildOpts := dockertypes.ImageBuildOptions{
		Tags:       []string{ref},
		Dockerfile: "Dockerfile",
		Remove:     true,
	}
	for k, v := range opts.Env {
		if strings.HasPrefix(k, "NEXT_PUBLIC_") {
			if buildOpts.BuildArgs == nil {
				buildOpts.BuildArgs = map[string]*string{}
			}
			val := v
			buildOpts.BuildArgs[k] = &val
		}
	}
	if cfg.BuildMemory > 0 {
		buildOpts.Memory = int64(cfg.BuildMemory) * 1024 * 1024
	}
	if cfg.RegistryAuth != "" {
		authStr, err := buildRegistryAuth(cfg.RegistryAuth, ref)
		if err != nil {
			return "", fmt.Errorf("failed to build registry auth for %s: %w", ref, err)
		}
		registryHost := strings.SplitN(ref, "/", 2)[0]
		buildOpts.AuthConfigs = map[string]registry.AuthConfig{
			registryHost: parseAuthConfig(authStr),
		}
	}

	buildResp, err := dockerCli.ImageBuild(ctx, buildCtx, buildOpts)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("docker build timed out after 20 minutes")
		}
		return "", fmt.Errorf("docker build failed: %w", err)
	}
	defer buildResp.Body.Close()

	// Drain docker build output, writing to the log file and surfacing errors.
	// Write directly (no bufio) so each line is flushed to disk immediately and
	// the log file tailer can stream build progress in real time.
	var logFile *os.File
	if logDir != "" {
		if err := os.MkdirAll(logDir, 0755); err == nil {
			logFile, _ = os.OpenFile(filepath.Join(logDir, strings.ToLower(svcName)+".log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if logFile != nil {
				defer logFile.Close()
			}
		}
	}

	scanner := bufio.NewScanner(buildResp.Body)
	for scanner.Scan() {
		var msg struct {
			Error  string `json:"error"`
			Stream string `json:"stream"`
		}
		if json.Unmarshal(scanner.Bytes(), &msg) != nil {
			continue
		}
		if msg.Error != "" {
			return "", fmt.Errorf("docker build: %s", strings.TrimSpace(msg.Error))
		}
		if line := normalizeDockerLine(msg.Stream); line != "" && logFile != nil {
			writeLogEntry(logFile, "INFO", line)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("docker build stream error: %w", err)
	}

	var authStr string
	if cfg.RegistryAuth != "" {
		authStr, err = buildRegistryAuth(cfg.RegistryAuth, ref)
		if err != nil {
			return "", fmt.Errorf("failed to build registry auth for push %s: %w", ref, err)
		}
	}
	pushRC, err := dockerCli.ImagePush(ctx, ref, image.PushOptions{RegistryAuth: authStr})
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("docker push timed out")
		}
		return "", fmt.Errorf("docker push failed: %w", err)
	}
	defer pushRC.Close()
	pushScanner := bufio.NewScanner(pushRC)
	for pushScanner.Scan() {
		var msg struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(pushScanner.Bytes(), &msg) == nil && msg.Error != "" {
			return "", fmt.Errorf("docker push failed: %s", strings.TrimSpace(msg.Error))
		}
	}
	if err := pushScanner.Err(); err != nil {
		return "", fmt.Errorf("docker push stream error: %w", err)
	}

	if _, err := dockerCli.ImageRemove(ctx, ref, image.RemoveOptions{Force: true}); err != nil {
		// Non-fatal: the image was pushed successfully; log and continue.
		_ = err
	}

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
		return "php:" + version + "-apache"
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

func generateDockerfile(opts BuildOpts) string {
	// builderImg is the primary FROM image. When BuilderImage is pre-resolved by
	// the version resolver it is used directly; otherwise we fall back to the
	// legacy runtime+version construction so that tests and direct callers that
	// only supply Runtime/Version continue to work.
	builderImg := opts.BuilderImage
	if builderImg == "" {
		builderImg = runtimeBaseImage(opts.runtime(), opts.version())
	}

	// runnerImg is the lightweight image used in the final stage of multi-stage
	// builds (Node.js, Java). It defaults to the builder image for runtimes that
	// use a single stage.
	runnerImg := func(fallback string) string {
		if opts.RunnerImage != "" {
			return opts.RunnerImage
		}
		return fallback
	}

	port := opts.Port
	if port == 0 {
		port = 8080
	}

	var b strings.Builder
	fmt.Fprintf(&b, "FROM %s\n\nWORKDIR /app\n\n", builderImg)

	// templateInstall is the install command already baked into each runtime's
	// template layer. We skip BuildCmd if it would just re-run the same thing.
	var templateInstall string

	switch opts.runtime() {
	case "nodejs":
		ver := opts.version()
		nodeAlpine := runnerImg(func() string {
			if ver == "" || ver == "lts" {
				return "node:lts-alpine"
			}
			return "node:" + ver + "-alpine"
		}())

		if opts.IsNextJS {
			for k := range opts.Env {
				if strings.HasPrefix(k, "NEXT_PUBLIC_") {
					fmt.Fprintf(&b, "ARG %s\n", k)
				}
			}
			b.WriteString("COPY package*.json ./\nRUN npm install\nCOPY . .\n")
			buildCmd := opts.BuildCmd
			if buildCmd == "" {
				buildCmd = "npm run build"
			}
			fmt.Fprintf(&b, "RUN %s\n", buildCmd)
			// Guarantee at least one next.config.* exists so the COPY in the runner never fails.
			b.WriteString("RUN [ -f next.config.js ] || [ -f next.config.mjs ] || echo 'module.exports = {};' > next.config.js\n")
			runCmd := opts.RunCmd
			if runCmd == "" {
				runCmd = "npm start"
			}
			fmt.Fprintf(&b, "\nFROM %s AS runner\nWORKDIR /app\n", nodeAlpine)
			b.WriteString("COPY package*.json ./\nRUN npm install --omit=dev\n")
			b.WriteString("COPY --from=0 /app/.next ./.next\n")
			b.WriteString("COPY --from=0 /app/public ./public\n")
			b.WriteString("COPY --from=0 /app/next.config.* ./\n")
			fmt.Fprintf(&b, "\nENV PORT=%d\nCMD [\"/bin/sh\", \"-c\", \"%s\"]\n", port, runCmd)
			return b.String()
		}
		if opts.BuildCmd != "" {
			// Multi-stage: full image to build, alpine to run.
			b.WriteString("COPY package*.json ./\nRUN npm install\nCOPY . .\n")
			fmt.Fprintf(&b, "RUN %s\n", opts.BuildCmd)
			fmt.Fprintf(&b, "\nFROM %s AS runner\nWORKDIR /app\n", nodeAlpine)
			b.WriteString("COPY package*.json ./\nRUN npm install --omit=dev\n")
			b.WriteString("COPY --from=0 /app ./\n")
			fmt.Fprintf(&b, "\nENV PORT=%d\n", port)
			if opts.RunCmd != "" {
				fmt.Fprintf(&b, "CMD %s\n", opts.RunCmd)
			}
			return b.String()
		}
		// No build step — install prod deps only on alpine directly.
		b.Reset()
		fmt.Fprintf(&b, "FROM %s\n\nWORKDIR /app\n\n", nodeAlpine)
		b.WriteString("COPY package*.json ./\nRUN npm install --omit=dev\nCOPY . .\n")
		templateInstall = "npm install --omit=dev"
	case "python":
		b.WriteString("COPY requirements.txt ./\n")
		// Auto-install system libs for common packages that need native extensions.
		// psycopg2 (not -binary) needs libpq-dev; mysqlclient needs libmysqlclient-dev;
		// Pillow needs image libs. Users on -binary variants or with custom deps use BuildCmd.
		b.WriteString("RUN apt-get update -qq && apt-get install -y --no-install-recommends \\\n")
		b.WriteString("    build-essential \\\n")
		b.WriteString("    $(grep -qiE '^psycopg2[^-]' requirements.txt 2>/dev/null && echo libpq-dev || true) \\\n")
		b.WriteString("    $(grep -qi 'mysqlclient' requirements.txt 2>/dev/null && echo default-libmysqlclient-dev || true) \\\n")
		b.WriteString("    $(grep -qiE '^Pillow|^pillow' requirements.txt 2>/dev/null && echo 'libjpeg-dev libpng-dev' || true) \\\n")
		b.WriteString("    && rm -rf /var/lib/apt/lists/*\n")
		b.WriteString("RUN pip install --no-cache-dir -r requirements.txt\nCOPY . .\n")
		templateInstall = "pip install --no-cache-dir -r requirements.txt"
	case "golang":
		b.WriteString("COPY go.mod ./\nRUN go mod download\nCOPY . .\n")
		if opts.BuildCmd != "" {
			fmt.Fprintf(&b, "RUN %s\n", opts.BuildCmd)
		} else {
			b.WriteString("RUN CGO_ENABLED=0 go build -o /bin/app .\n")
		}
		b.WriteString("\nFROM alpine:3\n")
		b.WriteString("RUN apk --no-cache add ca-certificates tzdata\n")
		b.WriteString("COPY --from=0 /bin/app /bin/app\n")
		fmt.Fprintf(&b, "\nENV PORT=%d\nCMD [\"/bin/app\"]\n", port)
		return b.String()
	case "php":
		b.Reset()
		fmt.Fprintf(&b, "FROM %s\n\n", builderImg)
		b.WriteString("RUN a2enmod rewrite\n")
		fmt.Fprintf(&b, "RUN sed -i 's/80/%d/g' /etc/apache2/sites-available/000-default.conf /etc/apache2/ports.conf\n\n", port)
		b.WriteString("WORKDIR /var/www/html\n\n")
		b.WriteString("COPY composer.* ./\n")
		// Install system libs and PHP extensions based on what composer.json declares.
		// pdo_pgsql needs libpq-dev; pdo_mysql is bundled; zip/gd are common Laravel deps.
		b.WriteString("RUN apt-get update -qq && apt-get install -y --no-install-recommends \\\n")
		b.WriteString("    unzip libzip-dev \\\n")
		b.WriteString("    $([ -f composer.json ] && grep -qiE 'pgsql|postgres' composer.json 2>/dev/null && echo libpq-dev || true) \\\n")
		b.WriteString("    && docker-php-ext-install zip \\\n")
		b.WriteString("    && ([ -f composer.json ] && grep -qiE 'pgsql|postgres' composer.json && docker-php-ext-install pdo_pgsql || true) \\\n")
		b.WriteString("    && ([ -f composer.json ] && grep -qiE 'mysql|mariadb' composer.json && docker-php-ext-install pdo_mysql || true) \\\n")
		b.WriteString("    && rm -rf /var/lib/apt/lists/*\n\n")
		b.WriteString("RUN if [ -f composer.json ]; then composer install --no-dev --optimize-autoloader; fi\n\n")
		b.WriteString("COPY . .\n\n")
		fmt.Fprintf(&b, "ENV PORT=%d\n", port)
		if opts.RunCmd != "" {
			fmt.Fprintf(&b, "CMD [\"/bin/sh\", \"-c\", \"%s\"]\n", opts.RunCmd)
		}
		return b.String()
	case "dotnet":
		sdkImg := builderImg
		aspnetImg := runnerImg(strings.Replace(builderImg, "/sdk:", "/aspnet:", 1))
		buildCmd := opts.BuildCmd
		if buildCmd == "" {
			buildCmd = "dotnet publish -c Release"
		}
		// Ensure a known output directory so the COPY in the runner stage is
		// always correct, even when the user omits -o from their build command.
		publishDir := dotnetPublishDir(buildCmd)
		if publishDir == "" {
			buildCmd += " -o out"
			publishDir = "out"
		}
		b.Reset()
		fmt.Fprintf(&b, "FROM %s AS build\n\nWORKDIR /app\n\n", sdkImg)
		b.WriteString("COPY *.csproj ./\nRUN dotnet restore\nCOPY . .\n")
		fmt.Fprintf(&b, "RUN %s\n", buildCmd)
		fmt.Fprintf(&b, "\nFROM %s\nWORKDIR /app\n", aspnetImg)
		fmt.Fprintf(&b, "COPY --from=build /app/%s .\n", publishDir)
		fmt.Fprintf(&b, "ENV ASPNETCORE_URLS=http://+:%d\n", port)
		runCmd := opts.RunCmd
		if runCmd == "" {
			runCmd = "f=$(find /app -maxdepth 1 -name '*.runtimeconfig.json' | head -1); dotnet ${f%.runtimeconfig.json}.dll"
		}
		fmt.Fprintf(&b, "CMD [\"/bin/sh\", \"-c\", \"%s\"]\n", runCmd)
		return b.String()
	case "ruby":
		b.WriteString("ENV RAILS_ENV=production RAILS_LOG_TO_STDOUT=1 RAILS_SERVE_STATIC_FILES=true\n")
		// Copy Gemfile first so we can detect the DB adapter before apt-get runs.
		b.WriteString("COPY Gemfile Gemfile.lock* ./\n")
		b.WriteString("RUN apt-get update -qq && apt-get install -y --no-install-recommends build-essential nodejs npm \\\n")
		b.WriteString("    && (grep -qE \"gem ['\\\"]pg['\\\"]\" Gemfile && apt-get install -y --no-install-recommends libpq-dev || true) \\\n")
		b.WriteString("    && (grep -qE \"gem ['\\\"]mysql2['\\\"]\" Gemfile && apt-get install -y --no-install-recommends default-libmysqlclient-dev || true) \\\n")
		b.WriteString("    && (grep -qE \"gem ['\\\"]sqlite3['\\\"]\" Gemfile && apt-get install -y --no-install-recommends libsqlite3-dev || true) \\\n")
		b.WriteString("    && npm install -g yarn --silent && rm -rf /var/lib/apt/lists/*\n")
		b.WriteString("RUN bundle config set --local without 'development test' && bundle install\n")
		b.WriteString("COPY . .\n")
		b.WriteString("RUN mkdir -p tmp/pids tmp/cache tmp/sockets log\n")
		if opts.BuildCmd != "" && strings.TrimSpace(opts.BuildCmd) != "bundle install" {
			fmt.Fprintf(&b, "RUN %s\n", opts.BuildCmd)
		}
		b.WriteString("RUN [ -f package.json ] && yarn install --frozen-lockfile 2>/dev/null || true\n")
		b.WriteString("RUN SECRET_KEY_BASE=placeholder bundle exec rails assets:precompile 2>/dev/null || true\n")
		runCmd := opts.RunCmd
		if runCmd == "" {
			runCmd = "bundle exec puma -C config/puma.rb"
		}
		fmt.Fprintf(&b, "\nENV PORT=%d\n", port)
		fmt.Fprintf(&b, "CMD [\"/bin/sh\", \"-c\", \"%s\"]\n", runCmd)
		return b.String()
	case "java":
		buildCmd := opts.BuildCmd
		if buildCmd == "" {
			buildCmd = "mvn package -DskipTests"
		}
		b.WriteString("COPY pom.xml ./\nRUN mvn dependency:go-offline -B\nCOPY . .\n")
		fmt.Fprintf(&b, "RUN %s\n", buildCmd)
		// Runner image: use pre-resolved RunnerImage when available, otherwise
		// derive from the resolved version tag stored in opts.Version.
		javaRunner := runnerImg(fmt.Sprintf("eclipse-temurin:%s-jre-alpine", opts.version()))
		fmt.Fprintf(&b, "\nFROM %s\nWORKDIR /app\n", javaRunner)
		b.WriteString("COPY --from=0 /app/target/*.jar app.jar\n")
		fmt.Fprintf(&b, "\nENV PORT=%d\nCMD [\"java\", \"-jar\", \"app.jar\"]\n", port)
		return b.String()
	default:
		b.WriteString("COPY . .\n")
	}

	if opts.BuildCmd != "" && strings.TrimSpace(opts.BuildCmd) != templateInstall {
		fmt.Fprintf(&b, "RUN %s\n", opts.BuildCmd)
	}
	fmt.Fprintf(&b, "\nENV PORT=%d\n", port)
	if opts.RunCmd != "" {
		fmt.Fprintf(&b, "CMD %s\n", opts.RunCmd)
	}
	return b.String()
}

// dotnetPublishDir extracts the output directory from a dotnet publish command.
// Returns "" when no -o / --output flag is present.
func dotnetPublishDir(cmd string) string {
	parts := strings.Fields(cmd)
	for i, p := range parts {
		switch p {
		case "-o", "--output":
			if i+1 < len(parts) {
				return strings.TrimRight(parts[i+1], "/")
			}
		default:
			if after, ok := strings.CutPrefix(p, "--output="); ok {
				return strings.TrimRight(after, "/")
			}
			if after, ok := strings.CutPrefix(p, "-o="); ok {
				return strings.TrimRight(after, "/")
			}
		}
	}
	return ""
}

func (o BuildOpts) runtime() string {
	if o.Runtime == "" {
		return "nodejs"
	}
	return o.Runtime
}

func (o BuildOpts) version() string {
	return o.Version
}

// buildRegistryAuth parses authB64 (base64-encoded JSON {"username","password"} or
// base64("user:pass") or a bare token) and returns a base64-encoded JSON auth string
// suitable for Docker SDK PullOptions.RegistryAuth / PushOptions.RegistryAuth.
func buildRegistryAuth(authB64, imageRef string) (string, error) {
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
			username, password = u, p
		} else {
			password = decoded
		}
	}

	registry := strings.SplitN(imageRef, "/", 2)[0]
	authCfg := struct {
		Username      string `json:"username"`
		Password      string `json:"password"`
		ServerAddress string `json:"serveraddress"`
	}{Username: username, Password: password, ServerAddress: registry}

	jsonBytes, err := json.Marshal(authCfg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth config: %w", err)
	}
	return base64.URLEncoding.EncodeToString(jsonBytes), nil
}

// parseAuthConfig decodes a buildRegistryAuth string back into a typed struct
// for use in ImageBuildOptions.AuthConfigs.
func parseAuthConfig(authStr string) registry.AuthConfig {
	raw, err := base64.URLEncoding.DecodeString(authStr)
	if err != nil {
		return registry.AuthConfig{}
	}
	var ac registry.AuthConfig
	json.Unmarshal(raw, &ac) //nolint:errcheck
	return ac
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
.bundle/
.rubygems.lock
/vendor/bundle/

# Build outputs
dist/
build/
out/
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
	dataDir := coreutils.GetDataDir()
	workDir := filepath.Join(dataDir, ".dployr", "services", coreutils.FormatName(name))
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

	// Always do a fresh clone into a clean directory. Re-using an existing
	// working directory via git pull is fragile (detached HEAD, partial state
	// from a failed previous build), so wipe and re-clone every time.
	if err := os.RemoveAll(destDir); err != nil {
		return fmt.Errorf("failed to clean destination directory: %s", err)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %s", err)
	}

	cloneCmd := fmt.Sprintf(
		"GIT_TERMINAL_PROMPT=0 git clone --depth 1 --branch %s %s .",
		remote.Branch, authUrl,
	)
	if err := shared.Exec(ctx, cloneCmd, destDir); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("git clone timed out after 5 minutes")
		}
		return fmt.Errorf("git clone failed: %s", err)
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
	return os.WriteFile(dockerIgnorePath, []byte(DockerIgnoreContent), 0644)
}

// PullImage pulls a docker image from a registry using the Docker SDK.
func PullImage(imageRef string, config *shared.Config, dockerCli deployDockerAPI) error {
	if !isValidDockerImageRef(imageRef) {
		return fmt.Errorf("invalid docker image reference: %s", imageRef)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var authStr string
	if config != nil && config.RegistryAuth != "" {
		var err error
		authStr, err = buildRegistryAuth(config.RegistryAuth, imageRef)
		if err != nil {
			return fmt.Errorf("registry auth failed: %w", err)
		}
	}

	rc, err := dockerCli.ImagePull(ctx, imageRef, image.PullOptions{RegistryAuth: authStr})
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("docker pull timed out after 5 minutes")
		}
		return fmt.Errorf("docker pull failed: %w", err)
	}
	defer rc.Close()
	if _, err := io.Copy(io.Discard, rc); err != nil {
		return fmt.Errorf("docker pull stream error: %w", err)
	}
	return nil
}

// DeployApp handles runtime setup, build, and service installation.
// Jobs (TypeJob) use the systemd/vfox bash path; all other types use the Go
// Docker path which avoids the bash script entirely.
func DeployApp(bp store.Blueprint, name, logPath string, cfg *shared.Config, dockerCli deployDockerAPI) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	if bp.Type == store.TypeStatic {
		// No container needed — the host Caddy proxy serves files directly.
		// registerProxyRoute in the worker registers the static.tpl route.
		return nil
	}

	if bp.Type == store.TypeJob {
		version := string(bp.Runtime.Version)
		if version == "" {
			return fmt.Errorf("runtime version cannot be empty")
		}
		return runDeployScript(ctx, bp, name, logPath, cfg)
	}

	return deployDocker(ctx, bp, name, logPath, cfg, dockerCli)
}

// deployDocker uses the Docker SDK to deploy web, worker, and static service
// types without spawning any shell processes.
func deployDocker(ctx context.Context, bp store.Blueprint, name, logPath string, cfg *shared.Config, dockerCli deployDockerAPI) error {
	port := bp.Port
	if port == 0 {
		port = 3000
	}

	// Write .env to disk so the process can source it at runtime (TypeJob reads it;
	// Docker containers also receive env vars inline via ContainerConfig.Env).
	if err := WriteEnvFile(bp.WorkingDir, bp, port); err != nil {
		return fmt.Errorf("failed to write env file: %w", err)
	}

	if bp.Image != "" {
		if err := PullImage(bp.Image, cfg, dockerCli); err != nil {
			return fmt.Errorf("failed to pull image: %w", err)
		}
	}

	cc := &ContainerConfig{
		Name:        name,
		Image:       bp.Image,
		Port:        port,
		HostPort:    coreutils.ComputeHostPort(name),
		Env:         buildEnv(bp, port),
		Description: bp.Desc,
		Type:        bp.Type,
		RunCmd:      bp.RunCmd,
	}
	if cfg != nil {
		cc.Memory = cfg.ContainerMemory
		cc.CPU = cfg.ContainerCPU
		cc.Storage = cfg.ContainerStorage
	}

	// Remove any pre-existing container with the same name (best-effort).
	dockerCli.ContainerRemove(ctx, name, container.RemoveOptions{Force: true}) //nolint:errcheck

	resp, err := dockerCli.ContainerCreate(ctx, ptr(cc.ContainerCfg()), ptr(cc.HostCfg()), nil, nil, name)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("docker create timed out")
		}
		return fmt.Errorf("docker create failed: %w", err)
	}

	if err := dockerCli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("docker start timed out")
		}
		return fmt.Errorf("docker start failed: %w", err)
	}

	shared.LogInfoF(name, logPath, fmt.Sprintf("container started: %s", resp.ID))
	return nil
}

// buildEnv builds a deduplicated KEY=value slice from a Blueprint.
// PORT is always first; EnvVars come next; Secrets fill in any remaining keys.
func buildEnv(bp store.Blueprint, port int) []string {
	var env []string
	written := map[string]bool{}
	write := func(k, v string) {
		if written[k] {
			return
		}
		written[k] = true
		env = append(env, k+"="+v)
	}
	write("PORT", fmt.Sprintf("%d", port))
	for k, v := range bp.EnvVars {
		write(k, v)
	}
	for k, v := range bp.Secrets {
		write(k, v)
	}
	return env
}

func ptr[T any](v T) *T { return &v }

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
		strconv.Itoa(coreutils.ComputeHostPort(bp.Name)),
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
		if err := scanner.Err(); err != nil {
			shared.LogErrF(name, logPath, fmt.Errorf("stdout read error: %w", err))
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
		if err := scanner.Err(); err != nil {
			shared.LogErrF(name, logPath, fmt.Errorf("stderr read error: %w", err))
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
