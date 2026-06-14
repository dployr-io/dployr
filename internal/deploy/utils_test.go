// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildAuthUrl_AlreadyHasCredentials(t *testing.T) {
	url := "https://user:pass@github.com/org/repo.git"
	got, err := buildAuthUrl(url, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != url {
		t.Errorf("buildAuthUrl(already credentialed) = %q, want unchanged", got)
	}
}

func TestBuildAuthUrl_EmptyToken(t *testing.T) {
	url := "https://github.com/org/repo.git"
	got, err := buildAuthUrl(url, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != url {
		t.Errorf("buildAuthUrl(empty token) = %q, want unchanged", got)
	}
}

func TestBuildAuthUrl_HttpNormalisedToHttps(t *testing.T) {
	got, err := buildAuthUrl("http://github.com/org/repo.git", "ghp_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(got, "https://") {
		t.Errorf("buildAuthUrl(http://) = %q, expected https:// prefix", got)
	}
	if !strings.Contains(got, "ghp_abc") {
		t.Errorf("buildAuthUrl(http://) = %q, expected token injected", got)
	}
}

func TestBuildAuthUrl_GitHub(t *testing.T) {
	got, err := buildAuthUrl("https://github.com/org/repo.git", "ghp_token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://x-access-token:ghp_token@github.com/org/repo.git"
	if got != want {
		t.Errorf("buildAuthUrl(github) = %q, want %q", got, want)
	}
}

func TestBuildAuthUrl_GitLab(t *testing.T) {
	got, err := buildAuthUrl("https://gitlab.com/org/repo.git", "glpat_token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "oauth2:glpat_token@") {
		t.Errorf("buildAuthUrl(gitlab) = %q, expected oauth2 username", got)
	}
}

func TestBuildAuthUrl_Bitbucket(t *testing.T) {
	got, err := buildAuthUrl("https://bitbucket.org/org/repo.git", "bb_token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "x-token-auth:bb_token@") {
		t.Errorf("buildAuthUrl(bitbucket) = %q, expected x-token-auth username", got)
	}
}

func TestBuildAuthUrl_UnknownHttpsHost(t *testing.T) {
	got, err := buildAuthUrl("https://git.internal.company/repo.git", "tok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "oauth2:tok@") {
		t.Errorf("buildAuthUrl(unknown host) = %q, expected oauth2 fallback", got)
	}
}

func TestBuildAuthUrl_SSHUrl(t *testing.T) {
	url := "git@github.com:org/repo.git"
	got, err := buildAuthUrl(url, "ghp_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != url {
		t.Errorf("buildAuthUrl(SSH) = %q, want unchanged %q", got, url)
	}
}

// fakeRegistry spins up a minimal Docker Registry v2 HTTP server that accepts
// any credentials and records the Authorization header it received.
func fakeRegistry(t *testing.T) (host string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return strings.TrimPrefix(srv.URL, "http://")
}

func b64JSON(username, password string) string {
	raw, _ := json.Marshal(map[string]string{"username": username, "password": password})
	return base64.StdEncoding.EncodeToString(raw)
}

func b64Colon(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

func decodeAuthStr(t *testing.T, authStr string) map[string]string {
	t.Helper()
	raw, err := base64.URLEncoding.DecodeString(authStr)
	if err != nil {
		t.Fatalf("base64 decode failed: %v", err)
	}
	var out map[string]string
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	return out
}

// TestBuildRegistryAuth_Base64JSONCreds covers auth stored as base64({"username":"…","password":"…"}).
func TestBuildRegistryAuth_Base64JSONCreds(t *testing.T) {
	host := fakeRegistry(t)
	auth := b64JSON("user@example.com", "secret-token")
	result, err := buildRegistryAuth(auth, host+"/apps:tag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := decodeAuthStr(t, result)
	if got["username"] != "user@example.com" {
		t.Errorf("username = %q, want user@example.com", got["username"])
	}
	if got["password"] != "secret-token" {
		t.Errorf("password = %q, want secret-token", got["password"])
	}
	if got["serveraddress"] != host {
		t.Errorf("serveraddress = %q, want %s", got["serveraddress"], host)
	}
}

// TestBuildRegistryAuth_Base64ColonCreds covers the standard Docker Basic Auth format: base64("user:pass").
func TestBuildRegistryAuth_Base64ColonCreds(t *testing.T) {
	host := fakeRegistry(t)
	auth := b64Colon("user@example.com", "dop_v1_sometoken")
	result, err := buildRegistryAuth(auth, host+"/apps:tag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := decodeAuthStr(t, result)
	if got["username"] != "user@example.com" {
		t.Errorf("username = %q, want user@example.com", got["username"])
	}
	if got["password"] != "dop_v1_sometoken" {
		t.Errorf("password = %q, want dop_v1_sometoken", got["password"])
	}
}

// TestBuildRegistryAuth_Base64RawToken covers base64 of a bare API token (no colon, no JSON).
func TestBuildRegistryAuth_Base64RawToken(t *testing.T) {
	host := fakeRegistry(t)
	auth := base64.StdEncoding.EncodeToString([]byte("dop_v1_sometoken"))
	result, err := buildRegistryAuth(auth, host+"/apps:tag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := decodeAuthStr(t, result)
	if got["password"] != "dop_v1_sometoken" {
		t.Errorf("password = %q, want dop_v1_sometoken", got["password"])
	}
}

// TestBuildRegistryAuth_RawToken covers a plain token stored directly (no base64 encoding).
func TestBuildRegistryAuth_RawToken(t *testing.T) {
	host := fakeRegistry(t)
	result, err := buildRegistryAuth("dop_v1_plaintexttoken", host+"/apps:tag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := decodeAuthStr(t, result)
	// A plain (non-base64) token falls through as the raw password value.
	if got["password"] == "" {
		t.Error("expected non-empty password for plain token input")
	}
}

func TestDetectNextJS_ConfigFile(t *testing.T) {
	for _, name := range []string{"next.config.js", "next.config.ts", "next.config.mjs"} {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, name), []byte("module.exports = {}"), 0644)
		if !detectNextJS(dir) {
			t.Errorf("detectNextJS: expected true for %s", name)
		}
	}
}

func TestDetectNextJS_PackageJSON(t *testing.T) {
	dir := t.TempDir()
	pkg := `{"dependencies":{"next":"14.0.0","react":"18.0.0"}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644)
	if !detectNextJS(dir) {
		t.Error("detectNextJS: expected true when next in dependencies")
	}
}

func TestDetectNextJS_NotNextJS(t *testing.T) {
	dir := t.TempDir()
	pkg := `{"dependencies":{"express":"4.0.0"}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644)
	if detectNextJS(dir) {
		t.Error("detectNextJS: expected false for plain express app")
	}
}

func TestGenerateDockerfile_NextJS(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "nodejs", Version: "20", BuildCmd: "npm run build", Port: 3000, IsNextJS: true})
	for _, want := range []string{
		"FROM node:20",
		"RUN npm install",
		"RUN npm run build",
		"FROM node:20-alpine AS runner",
		"npm install --omit=dev",
		"COPY --from=0 /app/.next ./.next",
		"COPY --from=0 /app/public ./public",
		"COPY --from=0 /app/next.config.* ./",
		`CMD ["/bin/sh", "-c", "npm start"]`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("generateDockerfile(nextjs): missing %q\n\ngot:\n%s", want, out)
		}
	}
	// Must NOT use standalone path
	if strings.Contains(out, ".next/standalone") {
		t.Errorf("generateDockerfile(nextjs): unexpected standalone path\n\ngot:\n%s", out)
	}
}

func TestGenerateDockerfile_NextJS_DefaultBuildCmd(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "nodejs", Version: "20", IsNextJS: true, Port: 3000})
	if !strings.Contains(out, "RUN npm run build") {
		t.Errorf("generateDockerfile(nextjs, no build_cmd): expected default npm run build\n\ngot:\n%s", out)
	}
}

func TestGenerateDockerfile_NextJS_DefaultRunCmd(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "nodejs", Version: "20", IsNextJS: true, Port: 3000})
	if !strings.Contains(out, "npm start") {
		t.Errorf("generateDockerfile(nextjs, no run_cmd): expected default npm start\n\ngot:\n%s", out)
	}
}

func TestGenerateDockerfile_NextJS_CustomRunCmd(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "nodejs", Version: "20", IsNextJS: true, Port: 3000, RunCmd: "node server.js"})
	if !strings.Contains(out, "node server.js") {
		t.Errorf("generateDockerfile(nextjs, custom run_cmd): expected custom run cmd\n\ngot:\n%s", out)
	}
	if strings.Contains(out, "npm start") {
		t.Errorf("generateDockerfile(nextjs, custom run_cmd): npm start should not appear when run_cmd is set\n\ngot:\n%s", out)
	}
}

func TestGenerateDockerfile_NextJS_ConfigCopied(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "nodejs", Version: "20", IsNextJS: true, Port: 3000})
	if !strings.Contains(out, "COPY --from=0 /app/next.config.* ./") {
		t.Errorf("generateDockerfile(nextjs): runner stage must copy next.config.*\n\ngot:\n%s", out)
	}
	// Fallback creation must be present in builder stage
	if !strings.Contains(out, "next.config.js") || !strings.Contains(out, "next.config.mjs") {
		t.Errorf("generateDockerfile(nextjs): builder must guarantee next.config.* exists\n\ngot:\n%s", out)
	}
}

func TestGenerateDockerfile_NextJS_NextPublicArgs(t *testing.T) {
	opts := BuildOpts{
		Runtime:  "nodejs",
		Version:  "20",
		IsNextJS: true,
		Port:     3000,
		Env: map[string]string{
			"NEXT_PUBLIC_API_URL": "https://api.example.com",
			"NEXT_PUBLIC_APP_ID":  "myapp",
			"SECRET_KEY":          "should-not-appear",
		},
	}
	out := generateDockerfile(opts)
	if !strings.Contains(out, "ARG NEXT_PUBLIC_API_URL") {
		t.Errorf("generateDockerfile(nextjs): missing ARG for NEXT_PUBLIC_API_URL\n\ngot:\n%s", out)
	}
	if !strings.Contains(out, "ARG NEXT_PUBLIC_APP_ID") {
		t.Errorf("generateDockerfile(nextjs): missing ARG for NEXT_PUBLIC_APP_ID\n\ngot:\n%s", out)
	}
	if strings.Contains(out, "ARG SECRET_KEY") {
		t.Errorf("generateDockerfile(nextjs): non-NEXT_PUBLIC_ var must not appear as ARG\n\ngot:\n%s", out)
	}
}

func TestGenerateDockerfile_NextJS_MultiStage(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "nodejs", Version: "20", IsNextJS: true, Port: 3000})
	stages := strings.Count(out, "FROM ")
	if stages < 2 {
		t.Errorf("generateDockerfile(nextjs): expected multi-stage build (>=2 FROM), got %d\n\ngot:\n%s", stages, out)
	}
	if !strings.Contains(out, "--omit=dev") {
		t.Errorf("generateDockerfile(nextjs): runner stage must use npm install --omit=dev\n\ngot:\n%s", out)
	}
}

func TestGenerateDockerfile_NodejsWithBuildCmd(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "nodejs", Version: "20", BuildCmd: "npm run build", RunCmd: "npm start", Port: 3000})
	for _, want := range []string{
		"FROM node:20", "RUN npm install", "RUN npm run build",
		"FROM node:20-alpine AS runner",
		"npm install --omit=dev",
		"CMD npm start", "ENV PORT=3000",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("generateDockerfile(nodejs+build): missing %q\n\ngot:\n%s", want, out)
		}
	}
}

func TestGenerateDockerfile_NodejsNoBuildCmd(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "nodejs", Version: "20", RunCmd: "node index.js", Port: 8080})
	for _, want := range []string{"FROM node:20-alpine", "npm install --omit=dev", "CMD node index.js", "ENV PORT=8080"} {
		if !strings.Contains(out, want) {
			t.Errorf("generateDockerfile(nodejs, no build_cmd): missing %q\n\ngot:\n%s", want, out)
		}
	}
	if strings.Contains(out, "RUN npm run") {
		t.Errorf("generateDockerfile(nodejs, no build_cmd): unexpected RUN build step\n\ngot:\n%s", out)
	}
}

func TestGenerateDockerfile_NodejsDefaultPort(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "nodejs", Version: "20"})
	if !strings.Contains(out, "FROM node:20-alpine") {
		t.Errorf("generateDockerfile(nodejs): expected alpine base\n\ngot:\n%s", out)
	}
	if !strings.Contains(out, "--omit=dev") {
		t.Errorf("generateDockerfile(nodejs): expected --omit=dev\n\ngot:\n%s", out)
	}
}

func TestGenerateDockerfile_Golang(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "golang", Version: "1.22", Port: 8080})
	for _, want := range []string{
		"FROM golang:1.22", "go mod download",
		"CGO_ENABLED=0 go build -o /bin/app .",
		"FROM alpine:3", "ca-certificates",
		"COPY --from=0 /bin/app /bin/app",
		`CMD ["/bin/app"]`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("generateDockerfile(golang): missing %q\n\ngot:\n%s", want, out)
		}
	}
	if strings.Contains(out, "golang:1.22\nWORKDIR") && strings.Contains(out, "CMD [\"/bin/app\"]") {
		// remove golang toolchain from final stage
		stages := strings.Split(out, "FROM alpine")
		if len(stages) < 2 {
			t.Errorf("generateDockerfile(golang): expected multi-stage build\n\ngot:\n%s", out)
		}
	}
}

func TestGenerateDockerfile_Python(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "python", Version: "3.12", RunCmd: "python app.py", Port: 5000})
	for _, want := range []string{"FROM python:3.12-slim", "pip install --no-cache-dir -r requirements.txt", "ENV PORT=5000"} {
		if !strings.Contains(out, want) {
			t.Errorf("generateDockerfile(python): missing %q\n\ngot:\n%s", want, out)
		}
	}
}

func TestGenerateDockerfile_Ruby(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "ruby", Version: "3.3", RunCmd: "ruby app.rb", Port: 4567})
	for _, want := range []string{"FROM ruby:3.3-slim", "bundle config set --local without", "bundle install", "ENV PORT=4567"} {
		if !strings.Contains(out, want) {
			t.Errorf("generateDockerfile(ruby): missing %q\n\ngot:\n%s", want, out)
		}
	}
}

func TestGenerateDockerfile_Ruby_DefaultRunCmd(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "ruby", Version: "3.3", Port: 8080})
	for _, want := range []string{
		"FROM ruby:3.3-slim",
		"bundle config set --local without",
		"mkdir -p tmp/pids",
		"SECRET_KEY_BASE=placeholder bundle exec rails assets:precompile",
		"bundle exec puma -C config/puma.rb",
		"ENV PORT=8080",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("generateDockerfile(ruby,default): missing %q\n\ngot:\n%s", want, out)
		}
	}
}

func TestGenerateDockerfile_Java(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "java", Version: "21", Port: 8080})
	for _, want := range []string{
		"FROM maven:3-eclipse-temurin-21",
		"mvn dependency:go-offline",
		"mvn package -DskipTests",
		"FROM eclipse-temurin:21-jre-alpine",
		"COPY --from=0 /app/target/*.jar app.jar",
		`CMD ["java", "-jar", "app.jar"]`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("generateDockerfile(java): missing %q\n\ngot:\n%s", want, out)
		}
	}
}

func TestGenerateDockerfile_DefaultPort(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "nodejs", Version: "20"})
	if !strings.Contains(out, "ENV PORT=8080") {
		t.Errorf("generateDockerfile: expected default PORT=8080\n\ngot:\n%s", out)
	}
}

func TestDotnetPublishDir_Default(t *testing.T) {
	if got := dotnetPublishDir("dotnet publish -c Release -o out"); got != "out" {
		t.Errorf("dotnetPublishDir default: got %q, want %q", got, "out")
	}
}

func TestDotnetPublishDir_CustomShortFlag(t *testing.T) {
	if got := dotnetPublishDir("dotnet publish -c Release -o publish"); got != "publish" {
		t.Errorf("dotnetPublishDir -o: got %q, want %q", got, "publish")
	}
}

func TestDotnetPublishDir_CustomLongFlag(t *testing.T) {
	if got := dotnetPublishDir("dotnet publish --output /app/dist"); got != "/app/dist" {
		t.Errorf("dotnetPublishDir --output: got %q, want %q", got, "/app/dist")
	}
}

func TestDotnetPublishDir_EqualsSyntax(t *testing.T) {
	if got := dotnetPublishDir("dotnet publish --output=release"); got != "release" {
		t.Errorf("dotnetPublishDir --output=: got %q, want %q", got, "release")
	}
}

func TestDotnetPublishDir_NoFlag(t *testing.T) {
	if got := dotnetPublishDir("dotnet publish -c Release"); got != "" {
		t.Errorf("dotnetPublishDir (no flag): got %q, want %q", got, "")
	}
}

func TestGenerateDockerfile_Dotnet(t *testing.T) {
	out := generateDockerfile(BuildOpts{
		Runtime:      "dotnet",
		BuilderImage: "mcr.microsoft.com/dotnet/sdk:9.0",
		RunnerImage:  "mcr.microsoft.com/dotnet/aspnet:9.0",
		Port:         3000,
	})
	for _, want := range []string{
		"FROM mcr.microsoft.com/dotnet/sdk:9.0 AS build",
		"COPY *.csproj ./",
		"RUN dotnet restore",
		"RUN dotnet publish -c Release -o out",
		"FROM mcr.microsoft.com/dotnet/aspnet:9.0",
		"COPY --from=build /app/out .",
		"ENV ASPNETCORE_URLS=http://+:3000",
		"runtimeconfig.json",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("generateDockerfile(dotnet): missing %q\n\ngot:\n%s", want, out)
		}
	}
}

func TestGenerateDockerfile_Dotnet_BuildCmdNoOutputFlag(t *testing.T) {
	out := generateDockerfile(BuildOpts{
		Runtime:      "dotnet",
		BuilderImage: "mcr.microsoft.com/dotnet/sdk:9.0",
		RunnerImage:  "mcr.microsoft.com/dotnet/aspnet:9.0",
		BuildCmd:     "dotnet publish -c Release",
		Port:         3000,
	})
	if !strings.Contains(out, "RUN dotnet publish -c Release -o out") {
		t.Errorf("generateDockerfile(dotnet, no -o): -o out should be injected\n\ngot:\n%s", out)
	}
	if !strings.Contains(out, "COPY --from=build /app/out .") {
		t.Errorf("generateDockerfile(dotnet, no -o): COPY should use injected out dir\n\ngot:\n%s", out)
	}
}

func TestGenerateDockerfile_Dotnet_CustomBuildCmd(t *testing.T) {
	out := generateDockerfile(BuildOpts{
		Runtime:      "dotnet",
		BuilderImage: "mcr.microsoft.com/dotnet/sdk:9.0",
		RunnerImage:  "mcr.microsoft.com/dotnet/aspnet:9.0",
		BuildCmd:     "dotnet publish -c Release -o publish",
		Port:         3000,
	})
	if !strings.Contains(out, "RUN dotnet publish -c Release -o publish") {
		t.Errorf("generateDockerfile(dotnet, custom build): missing custom build cmd\n\ngot:\n%s", out)
	}
	if !strings.Contains(out, "COPY --from=build /app/publish .") {
		t.Errorf("generateDockerfile(dotnet, custom build): COPY should use custom output dir\n\ngot:\n%s", out)
	}
}

func TestGenerateDockerfile_Dotnet_CustomRunCmd(t *testing.T) {
	out := generateDockerfile(BuildOpts{
		Runtime:      "dotnet",
		BuilderImage: "mcr.microsoft.com/dotnet/sdk:9.0",
		RunnerImage:  "mcr.microsoft.com/dotnet/aspnet:9.0",
		RunCmd:       "dotnet MyApp.dll",
		Port:         3000,
	})
	if !strings.Contains(out, "dotnet MyApp.dll") {
		t.Errorf("generateDockerfile(dotnet, custom run): missing custom run cmd\n\ngot:\n%s", out)
	}
	if strings.Contains(out, "runtimeconfig.json") {
		t.Errorf("generateDockerfile(dotnet, custom run): find fallback should not appear\n\ngot:\n%s", out)
	}
}

func TestGenerateDockerfile_Dotnet_MultiStage(t *testing.T) {
	out := generateDockerfile(BuildOpts{
		Runtime:      "dotnet",
		BuilderImage: "mcr.microsoft.com/dotnet/sdk:9.0",
		RunnerImage:  "mcr.microsoft.com/dotnet/aspnet:9.0",
		Port:         3000,
	})
	if stages := strings.Count(out, "FROM "); stages < 2 {
		t.Errorf("generateDockerfile(dotnet): expected multi-stage build (>=2 FROM), got %d\n\ngot:\n%s", stages, out)
	}
}

func TestEnsureDockerfile_WritesWhenMissing(t *testing.T) {
	dir := t.TempDir()
	opts := BuildOpts{Runtime: "nodejs", Version: "20", RunCmd: "npm start", Port: 3000}
	if err := ensureDockerfile(dir, opts); err != nil {
		t.Fatalf("ensureDockerfile failed: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if err != nil {
		t.Fatalf("Dockerfile not written: %v", err)
	}
	if !strings.Contains(string(content), "FROM node:20") {
		t.Errorf("written Dockerfile missing expected content\n\ngot:\n%s", content)
	}
}

func TestEnsureDockerfile_OverwritesStaleFile(t *testing.T) {
	// A Dockerfile that exists on disk but is NOT tracked by git (e.g. left by a
	// previous failed build) must be overwritten with a freshly generated one.
	dir := t.TempDir()
	stale := []byte("\n") // 1-byte stale file, as seen in production
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), stale, 0644); err != nil {
		t.Fatal(err)
	}
	if err := ensureDockerfile(dir, BuildOpts{Runtime: "nodejs", Version: "20", RunCmd: "npm start"}); err != nil {
		t.Fatalf("ensureDockerfile failed: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if err != nil {
		t.Fatalf("Dockerfile missing after ensureDockerfile: %v", err)
	}
	if !strings.Contains(string(content), "FROM node:20") {
		t.Errorf("stale Dockerfile was not overwritten\n\ngot:\n%s", content)
	}
}

func TestEnsureDockerfile_PreservesCommitted(t *testing.T) {
	// A Dockerfile tracked by git must not be overwritten.
	dir := t.TempDir()

	// Init a git repo and commit a Dockerfile so ls-files reports it as tracked.
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "test"},
	} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	committed := "FROM scratch\nCMD [\"custom\"]\n"
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(committed), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "Dockerfile"}, {"commit", "-m", "add dockerfile"}} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	if err := ensureDockerfile(dir, BuildOpts{Runtime: "nodejs", Version: "20"}); err != nil {
		t.Fatalf("ensureDockerfile failed: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != committed {
		t.Errorf("ensureDockerfile overwrote committed Dockerfile\n\ngot:\n%s", content)
	}
}
