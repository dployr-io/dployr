// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

// TestRegistryLogin_Base64JSONCreds covers auth stored as base64({"username":"…","password":"…"}).
func TestRegistryLogin_Base64JSONCreds(t *testing.T) {
	host := fakeRegistry(t)
	auth := b64JSON("user@example.com", "secret-token")
	if err := registryLogin(context.Background(), host, auth, t.TempDir()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRegistryLogin_Base64ColonCreds covers the standard Docker Basic Auth format: base64("user:pass").
func TestRegistryLogin_Base64ColonCreds(t *testing.T) {
	host := fakeRegistry(t)
	auth := b64Colon("user@example.com", "dop_v1_sometoken")
	if err := registryLogin(context.Background(), host, auth, t.TempDir()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRegistryLogin_Base64RawToken covers base64 of a bare API token (no colon, no JSON).
func TestRegistryLogin_Base64RawToken(t *testing.T) {
	host := fakeRegistry(t)
	auth := base64.StdEncoding.EncodeToString([]byte("dop_v1_sometoken"))
	if err := registryLogin(context.Background(), host, auth, t.TempDir()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRegistryLogin_RawToken covers a plain token stored directly (no base64 encoding).
func TestRegistryLogin_RawToken(t *testing.T) {
	host := fakeRegistry(t)
	if err := registryLogin(context.Background(), host, "dop_v1_plaintexttoken", t.TempDir()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateDockerfile_NodejsWithBuildCmd(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "nodejs", Version: "20", BuildCmd: "npm run build", RunCmd: "npm start", Port: 3000})
	for _, want := range []string{"FROM node:20", "RUN npm install", "RUN npm run build", "CMD npm start", "ENV PORT=3000"} {
		if !strings.Contains(out, want) {
			t.Errorf("generateDockerfile(nodejs): missing %q\n\ngot:\n%s", want, out)
		}
	}
}

func TestGenerateDockerfile_NodejsNoBuildCmd(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "nodejs", Version: "20", RunCmd: "node index.js", Port: 8080})
	if strings.Contains(out, "RUN npm run") {
		t.Errorf("generateDockerfile(nodejs, no build_cmd): unexpected RUN build step\n\ngot:\n%s", out)
	}
	if !strings.Contains(out, "CMD node index.js") {
		t.Errorf("generateDockerfile(nodejs, no build_cmd): missing CMD\n\ngot:\n%s", out)
	}
}

func TestGenerateDockerfile_Golang(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "golang", Version: "1.22", Port: 8080})
	for _, want := range []string{"FROM golang:1.22", "go mod download", "go build -o /app/bin", "CMD [\"/app/bin\"]"} {
		if !strings.Contains(out, want) {
			t.Errorf("generateDockerfile(golang): missing %q\n\ngot:\n%s", want, out)
		}
	}
}

func TestGenerateDockerfile_Python(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "python", Version: "3.12", RunCmd: "python app.py", Port: 5000})
	for _, want := range []string{"FROM python:3.12", "pip install -r requirements.txt", "ENV PORT=5000"} {
		if !strings.Contains(out, want) {
			t.Errorf("generateDockerfile(python): missing %q\n\ngot:\n%s", want, out)
		}
	}
}

func TestGenerateDockerfile_DefaultPort(t *testing.T) {
	out := generateDockerfile(BuildOpts{Runtime: "nodejs", Version: "20"})
	if !strings.Contains(out, "ENV PORT=8080") {
		t.Errorf("generateDockerfile: expected default PORT=8080\n\ngot:\n%s", out)
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

func TestEnsureDockerfile_PreservesExisting(t *testing.T) {
	dir := t.TempDir()
	existing := "FROM scratch\nCMD [\"custom\"]\n"
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}
	if err := ensureDockerfile(dir, BuildOpts{Runtime: "nodejs", Version: "20"}); err != nil {
		t.Fatalf("ensureDockerfile failed: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != existing {
		t.Errorf("ensureDockerfile overwrote existing Dockerfile\n\ngot:\n%s", content)
	}
}
