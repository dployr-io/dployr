// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dployr-io/dployr/pkg/store"
)

func TestWriteEnvFile_Basic(t *testing.T) {
	dir := t.TempDir()
	bp := store.Blueprint{
		EnvVars: map[string]string{"APP_ENV": "production", "DB_URL": "postgres://localhost/app"},
	}
	if err := WriteEnvFile(dir, bp, 3000); err != nil {
		t.Fatalf("WriteEnvFile: %v", err)
	}
	content := mustReadEnvFile(t, dir)
	if !strings.Contains(content, "PORT=3000\n") {
		t.Errorf("missing PORT=3000\ngot:\n%s", content)
	}
	if !strings.Contains(content, "APP_ENV=production\n") {
		t.Errorf("missing APP_ENV=production\ngot:\n%s", content)
	}
	if !strings.Contains(content, "DB_URL=postgres://localhost/app\n") {
		t.Errorf("missing DB_URL\ngot:\n%s", content)
	}
}

func TestWriteEnvFile_EmptyBlueprint(t *testing.T) {
	dir := t.TempDir()
	if err := WriteEnvFile(dir, store.Blueprint{}, 8080); err != nil {
		t.Fatalf("WriteEnvFile: %v", err)
	}
	if got := mustReadEnvFile(t, dir); got != "PORT=8080\n" {
		t.Errorf("expected only PORT=8080, got:\n%s", got)
	}
}

func TestWriteEnvFile_PortNotOverriddenByEnvVars(t *testing.T) {
	dir := t.TempDir()
	bp := store.Blueprint{
		EnvVars: map[string]string{"PORT": "9999"},
	}
	if err := WriteEnvFile(dir, bp, 3000); err != nil {
		t.Fatalf("WriteEnvFile: %v", err)
	}
	content := mustReadEnvFile(t, dir)
	if !strings.HasPrefix(content, "PORT=3000\n") {
		t.Errorf("configured port must appear first; got:\n%s", content)
	}
	if strings.Contains(content, "PORT=9999") {
		t.Errorf("env var PORT=9999 must not override configured port\ngot:\n%s", content)
	}
}

func TestWriteEnvFile_SecretsDontDuplicateEnvVars(t *testing.T) {
	dir := t.TempDir()
	bp := store.Blueprint{
		EnvVars: map[string]string{"API_KEY": "from-env"},
		Secrets: map[string]string{"API_KEY": "from-secret"},
	}
	if err := WriteEnvFile(dir, bp, 3000); err != nil {
		t.Fatalf("WriteEnvFile: %v", err)
	}
	content := mustReadEnvFile(t, dir)
	if count := strings.Count(content, "API_KEY="); count != 1 {
		t.Errorf("API_KEY written %d times, want 1\ngot:\n%s", count, content)
	}
}

func TestWriteEnvFile_SecretsWrittenWhenNoEnvVarConflict(t *testing.T) {
	dir := t.TempDir()
	bp := store.Blueprint{
		Secrets: map[string]string{"DB_PASS": "hunter2"},
	}
	if err := WriteEnvFile(dir, bp, 3000); err != nil {
		t.Fatalf("WriteEnvFile: %v", err)
	}
	if !strings.Contains(mustReadEnvFile(t, dir), "DB_PASS=hunter2\n") {
		t.Error("secret DB_PASS not written to .env")
	}
}

func TestWriteEnvFile_FilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits are not enforced on Windows")
	}
	dir := t.TempDir()
	if err := WriteEnvFile(dir, store.Blueprint{}, 3000); err != nil {
		t.Fatalf("WriteEnvFile: %v", err)
	}
	info, err := os.Stat(filepath.Join(dir, ".env"))
	if err != nil {
		t.Fatalf("stat .env: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf(".env permissions = %04o, want 0600", perm)
	}
}

func mustReadEnvFile(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".env"))
	if err != nil {
		t.Fatalf("read .env: %v", err)
	}
	return string(data)
}
