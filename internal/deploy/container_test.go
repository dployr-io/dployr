// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"slices"
	"testing"
	"time"

	"github.com/dployr-io/dployr/pkg/store"
)

func TestContainerConfig_BasicWeb(t *testing.T) {
	cfg := &ContainerConfig{
		Name:     "my-app",
		Image:    "registry/my-app:latest",
		Port:     3000,
		HostPort: 62000,
		Env:      []string{"PORT=3000", "APP_ENV=production"},
		Type:     store.TypeWeb,
	}

	cc := cfg.ContainerCfg()
	hc := cfg.HostCfg()

	if cc.Image != "registry/my-app:latest" {
		t.Errorf("image = %q, want registry/my-app:latest", cc.Image)
	}
	if !slices.Contains(cc.Env, "PORT=3000") {
		t.Errorf("Env missing PORT=3000, got %v", cc.Env)
	}
	if _, ok := cc.ExposedPorts["3000/tcp"]; !ok {
		t.Error("ExposedPorts missing 3000/tcp")
	}
	if hc.RestartPolicy.Name != "unless-stopped" {
		t.Errorf("RestartPolicy = %q, want unless-stopped", hc.RestartPolicy.Name)
	}
	bindings := hc.PortBindings["3000/tcp"]
	if len(bindings) == 0 || bindings[0].HostPort != "62000" {
		t.Errorf("PortBindings[3000/tcp] = %v, want HostPort 62000", bindings)
	}
}

func TestContainerConfig_WithRunCmd(t *testing.T) {
	cfg := &ContainerConfig{
		Name:   "worker",
		Image:  "registry/worker:v1",
		Type:   store.TypeWorker,
		RunCmd: "node worker.js",
	}
	cc := cfg.ContainerCfg()
	if len(cc.Entrypoint) == 0 || cc.Entrypoint[0] != "/bin/sh" {
		t.Errorf("Entrypoint = %v, want [/bin/sh]", cc.Entrypoint)
	}
	if len(cc.Cmd) < 2 || cc.Cmd[0] != "-c" || cc.Cmd[1] != "node worker.js" {
		t.Errorf("Cmd = %v, want [-c, node worker.js]", cc.Cmd)
	}
}

func TestContainerConfig_NoPortBindingWhenZero(t *testing.T) {
	cfg := &ContainerConfig{Name: "w", Image: "img", Type: store.TypeWorker}
	hc := cfg.HostCfg()
	if len(hc.PortBindings) != 0 {
		t.Errorf("PortBindings should be empty for zero port, got %v", hc.PortBindings)
	}
}

func TestContainerConfig_HealthCheck(t *testing.T) {
	cfg := &ContainerConfig{
		Name:  "api",
		Image: "registry/api:v1",
		Port:  8080,
		Type:  store.TypeWeb,
		HealthCheck: &store.HealthCheck{
			Path:     "/healthz",
			Interval: 15,
			Timeout:  3,
			Retries:  5,
		},
	}
	cc := cfg.ContainerCfg()
	hc := cc.Healthcheck
	if hc == nil {
		t.Fatal("Healthcheck is nil")
	}
	if len(hc.Test) < 2 || hc.Test[1] != "curl -sf http://localhost:8080/healthz || exit 1" {
		t.Errorf("Healthcheck.Test = %v", hc.Test)
	}
	if hc.Interval != 15*time.Second {
		t.Errorf("Interval = %v, want 15s", hc.Interval)
	}
	if hc.Timeout != 3*time.Second {
		t.Errorf("Timeout = %v, want 3s", hc.Timeout)
	}
	if hc.Retries != 5 {
		t.Errorf("Retries = %d, want 5", hc.Retries)
	}
}

func TestContainerConfig_HealthCheckDefaults(t *testing.T) {
	cfg := &ContainerConfig{
		Name:        "api",
		Image:       "registry/api:v1",
		Port:        8080,
		Type:        store.TypeWeb,
		HealthCheck: &store.HealthCheck{Path: "/health"},
	}
	hc := cfg.ContainerCfg().Healthcheck
	if hc == nil {
		t.Fatal("Healthcheck is nil")
	}
	if hc.Interval != 30*time.Second {
		t.Errorf("default Interval = %v, want 30s", hc.Interval)
	}
	if hc.Timeout != 5*time.Second {
		t.Errorf("default Timeout = %v, want 5s", hc.Timeout)
	}
	if hc.Retries != 3 {
		t.Errorf("default Retries = %d, want 3", hc.Retries)
	}
}

func TestContainerConfig_HealthCheckSkippedWhenNoPath(t *testing.T) {
	cfg := &ContainerConfig{
		Name:        "api",
		Image:       "img",
		Type:        store.TypeWeb,
		HealthCheck: &store.HealthCheck{}, // Path is empty
	}
	if cfg.ContainerCfg().Healthcheck != nil {
		t.Error("Healthcheck must not be added when HealthCheck.Path is empty")
	}
}

func TestContainerConfig_ResourceLimits(t *testing.T) {
	cfg := &ContainerConfig{
		Name:    "limited",
		Image:   "registry/app:v1",
		Type:    store.TypeWeb,
		Memory:  512,
		CPU:     500,
		Storage: 10,
	}
	hc := cfg.HostCfg()

	wantMem := int64(512 * 1024 * 1024)
	if hc.Resources.Memory != wantMem {
		t.Errorf("Memory = %d, want %d", hc.Resources.Memory, wantMem)
	}
	if hc.Resources.MemorySwap != wantMem {
		t.Errorf("MemorySwap = %d, want %d", hc.Resources.MemorySwap, wantMem)
	}
	if hc.Resources.CPUQuota != 50000 {
		t.Errorf("CPUQuota = %d, want 50000 (500 millicores)", hc.Resources.CPUQuota)
	}
	if hc.StorageOpt["size"] != "10g" {
		t.Errorf("StorageOpt[size] = %q, want 10g", hc.StorageOpt["size"])
	}
}

func TestContainerConfig_NoResourceFlagsWhenZero(t *testing.T) {
	cfg := &ContainerConfig{Name: "app", Image: "img", Type: store.TypeWeb}
	hc := cfg.HostCfg()
	if hc.Resources.Memory != 0 {
		t.Errorf("Memory should be 0 when not set, got %d", hc.Resources.Memory)
	}
	if hc.Resources.CPUQuota != 0 {
		t.Errorf("CPUQuota should be 0 when not set, got %d", hc.Resources.CPUQuota)
	}
	if len(hc.StorageOpt) != 0 {
		t.Errorf("StorageOpt should be empty when not set, got %v", hc.StorageOpt)
	}
}

func TestContainerConfig_Description(t *testing.T) {
	cfg := &ContainerConfig{
		Name:        "app",
		Image:       "registry/app:v1",
		Type:        store.TypeWeb,
		Description: "My Production App",
	}
	if cfg.ContainerCfg().Labels["description"] != "My Production App" {
		t.Errorf("Labels[description] = %q, want My Production App", cfg.ContainerCfg().Labels["description"])
	}
}

func TestContainerConfig_NoLabelWhenEmpty(t *testing.T) {
	cfg := &ContainerConfig{Name: "app", Image: "img", Type: store.TypeWeb}
	labels := cfg.ContainerCfg().Labels
	if v, ok := labels["description"]; ok && v != "" {
		t.Errorf("unexpected description label %q when Description is empty", v)
	}
}

func TestResolveStaticDir_Empty(t *testing.T) {
	if got := ResolveStaticDir("/workdir", ""); got != "/workdir" {
		t.Errorf("got %q, want /workdir", got)
	}
}

func TestResolveStaticDir_Relative(t *testing.T) {
	got := ResolveStaticDir("/workdir", "dist")
	if got != "/workdir/dist" {
		t.Errorf("got %q, want /workdir/dist", got)
	}
}

func TestResolveStaticDir_Absolute(t *testing.T) {
	got := ResolveStaticDir("/workdir", "/custom/path")
	if got != "/custom/path" {
		t.Errorf("got %q, want /custom/path (unchanged)", got)
	}
}
