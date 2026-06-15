// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"fmt"
	"path"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"

	"github.com/dployr-io/dployr/pkg/store"
)

// ContainerConfig holds all parameters needed to create a Docker container.
// Build it from a Blueprint then call ContainerCfg() / HostCfg() to get the
// typed SDK structs — no shell invocations, fully testable.
type ContainerConfig struct {
	Name        string
	Image       string
	Port        int      // container port; 0 skips port binding
	HostPort    int      // host port; 0 skips port binding
	Env         []string // KEY=value entries passed inline to the container
	Description string
	Type        store.ServiceType
	RunCmd      string // optional CMD override
	Memory      int    // MB; 0 = no limit
	CPU         int    // millicores; 0 = no limit
	Storage     int    // GB; 0 = no limit
	ClusterID   string // when set, container is placed in the cluster's cgroup slice
}

// ContainerCfg returns the container.Config for docker ContainerCreate.
func (c *ContainerConfig) ContainerCfg() container.Config {
	cfg := container.Config{
		Image:  c.Image,
		Env:    c.Env,
		Labels: map[string]string{},
	}

	if c.Description != "" {
		cfg.Labels["description"] = c.Description
	}

	if c.Port > 0 {
		cfg.ExposedPorts = nat.PortSet{
			nat.Port(fmt.Sprintf("%d/tcp", c.Port)): struct{}{},
		}
	}

	if c.RunCmd != "" {
		cfg.Entrypoint = []string{"/bin/sh"}
		cfg.Cmd = []string{"-c", c.RunCmd}
	}

	return cfg
}

// HostCfg returns the container.HostConfig for docker ContainerCreate.
func (c *ContainerConfig) HostCfg() container.HostConfig {
	hc := container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
	}

	if c.Port > 0 && c.HostPort > 0 {
		hc.PortBindings = nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", c.Port)): []nat.PortBinding{
				{HostPort: fmt.Sprintf("%d", c.HostPort)},
			},
		}
	}

	if c.Memory > 0 {
		memBytes := int64(c.Memory) * 1024 * 1024
		hc.Resources.Memory = memBytes
		hc.Resources.MemorySwap = memBytes
	}
	if c.CPU > 0 {
		// CPUQuota µs per 100ms period: 500 millicores → 50000 µs
		hc.Resources.CPUQuota = int64(c.CPU) * 100
	}
	if c.Storage > 0 {
		hc.StorageOpt = map[string]string{"size": fmt.Sprintf("%dg", c.Storage)}
	}
	if c.ClusterID != "" {
		hc.CgroupParent = "dployr-cluster-" + c.ClusterID + ".slice"
	}

	return hc
}

// resolveStaticDir returns the absolute host path for the static content directory.
// Relative paths are joined with workDir; absolute paths are returned unchanged.
// Empty staticDir returns workDir itself.
func ResolveStaticDir(workDir, staticDir string) string {
	if staticDir == "" {
		return workDir
	}
	// Use path.IsAbs (POSIX) — these are Linux server paths, not Windows paths.
	if path.IsAbs(staticDir) {
		return staticDir
	}
	return path.Join(workDir, staticDir)
}
