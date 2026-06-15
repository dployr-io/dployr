// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

//go:build linux

package system

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	cgroup2 "github.com/containerd/cgroups/v3/cgroup2"
	systemddbus "github.com/coreos/go-systemd/v22/dbus"
	"github.com/dployr-io/dployr/pkg/core/system"
)

const cgroupRoot = "/sys/fs/cgroup"

// EnsureClusterSlice creates or updates the per-cluster systemd slice with
// the given memory (MB) and CPU (millicores) limits. It is idempotent —
// calling it repeatedly with the same values is a no-op, and calling it with
// changed values (tier upgrade/downgrade) updates the slice in place.
func EnsureClusterSlice(clusterID string, memoryMB int, cpuMillicores int) error {
	sliceName := "dployr-cluster-" + clusterID + ".slice"
	sliceFile := "/etc/systemd/system/" + sliceName

	var memLine, cpuLine string
	if memoryMB > 0 {
		memLine = fmt.Sprintf("MemoryMax=%dM", memoryMB)
	}
	if cpuMillicores > 0 {
		cpuLine = fmt.Sprintf("CPUQuota=%d%%", cpuMillicores/10)
	}

	desired := fmt.Sprintf("[Unit]\nDescription=dployr cluster %s\nBefore=slices.target\n\n[Slice]\n%s\n%s\n",
		clusterID, memLine, cpuLine)

	current, _ := os.ReadFile(sliceFile)
	if string(current) == desired {
		return nil // already up to date
	}

	if err := os.WriteFile(sliceFile, []byte(desired), 0644); err != nil {
		return fmt.Errorf("write slice unit %s: %w", sliceName, err)
	}

	conn, err := systemddbus.NewSystemConnectionContext(context.TODO())
	if err != nil {
		return fmt.Errorf("connect to systemd dbus: %w", err)
	}
	defer conn.Close()

	if err := conn.ReloadContext(context.TODO()); err != nil {
		return fmt.Errorf("systemd daemon-reload: %w", err)
	}

	ch := make(chan string, 1)
	if _, err := conn.StartUnitContext(context.TODO(), sliceName, "replace", ch); err != nil {
		return fmt.Errorf("start slice %s: %w", sliceName, err)
	}
	<-ch

	return nil
}

// ReadClusterResources returns per-cluster cgroup v2 memory stats keyed by
// cluster ID. Returns nil when no cluster slices exist on this node.
func ReadClusterResources() map[string]*system.ClusterResourcesInfo {
	// Cluster slices sit two levels deep under dployr.slice due to systemd's
	// hyphen-as-hierarchy naming: dployr-cluster-<id>.slice →
	//   /sys/fs/cgroup/dployr.slice/dployr-cluster.slice/dployr-cluster-<id>.slice/
	pattern := filepath.Join(cgroupRoot, "dployr.slice", "dployr-cluster.slice", "dployr-cluster-*.slice")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return nil
	}

	result := make(map[string]*system.ClusterResourcesInfo, len(matches))
	for _, slicePath := range matches {
		base := filepath.Base(slicePath)
		id := strings.TrimSuffix(strings.TrimPrefix(base, "dployr-cluster-"), ".slice")
		if id == "" {
			continue
		}

		// cgroupPath is relative to the cgroup root, as required by cgroup2.Load.
		cgroupPath := strings.TrimPrefix(slicePath, cgroupRoot)
		m, err := cgroup2.Load(cgroupPath)
		if err != nil {
			continue
		}

		stats, err := m.Stat()
		if err != nil || stats.Memory == nil {
			continue
		}

		// UsageLimit is math.MaxUint64 when memory.max = "max" (unlimited).
		limit := int64(stats.Memory.UsageLimit)
		if stats.Memory.UsageLimit == math.MaxUint64 {
			limit = 0
		}

		result[id] = &system.ClusterResourcesInfo{
			MemoryUsedBytes:  int64(stats.Memory.Usage),
			MemoryLimitBytes: limit,
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}
