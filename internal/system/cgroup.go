// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

//go:build linux

package system

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"strings"

	cgroup2 "github.com/containerd/cgroups/v3/cgroup2"
	systemddbus "github.com/coreos/go-systemd/v22/dbus"
	"github.com/dployr-io/dployr/pkg/core/system"
	godbus "github.com/godbus/dbus/v5"
)

const cgroupRoot = "/sys/fs/cgroup"

// EnsureClusterSlice creates or updates the per-cluster systemd slice with
// the given memory (MB) and CPU (millicores) limits via D-Bus transient units —
// no file writes required. It is idempotent: if the slice is already active
// with the correct limits it is a no-op; if limits differ they are updated
// in-place via SetUnitProperties.
func EnsureClusterSlice(clusterID string, memoryMB int, cpuMillicores int) error {
	sliceName := "dployr-cluster-" + clusterID + ".slice"

	// Desired values in D-Bus wire types.
	// MemoryMax: bytes (uint64); math.MaxUint64 = "max" (unlimited).
	// CPUQuotaPerSecUSec: µs per second; math.MaxUint64 = unlimited.
	var wantMem uint64 = math.MaxUint64
	if memoryMB > 0 {
		wantMem = uint64(memoryMB) * 1024 * 1024
	}
	var wantCPU uint64 = math.MaxUint64
	if cpuMillicores > 0 {
		// 1000 millicores == 1 CPU == 1_000_000 µs/s
		wantCPU = uint64(cpuMillicores) * 1000
	}

	ctx := context.TODO()
	conn, err := systemddbus.NewSystemConnectionContext(ctx)
	if err != nil {
		return fmt.Errorf("connect to systemd dbus: %w", err)
	}
	defer conn.Close()

	// Check whether the slice is already loaded with the correct limits.
	existing, err := conn.GetUnitTypePropertiesContext(ctx, sliceName, "Slice")
	if err == nil {
		curMem, _ := existing["MemoryMax"].(uint64)
		curCPU, _ := existing["CPUQuotaPerSecUSec"].(uint64)
		if curMem == wantMem && curCPU == wantCPU {
			return nil // already correct, nothing to do
		}
		// Slice exists but limits differ — update in place.
		props := []systemddbus.Property{
			{Name: "MemoryMax", Value: godbus.MakeVariant(wantMem)},
			{Name: "CPUQuotaPerSecUSec", Value: godbus.MakeVariant(wantCPU)},
		}
		return conn.SetUnitPropertiesContext(ctx, sliceName, true, props...)
	}

	// Slice does not exist yet — create it as a transient unit.
	props := []systemddbus.Property{
		systemddbus.PropDescription("dployr cluster " + clusterID),
		{Name: "MemoryMax", Value: godbus.MakeVariant(wantMem)},
		{Name: "CPUQuotaPerSecUSec", Value: godbus.MakeVariant(wantCPU)},
	}
	ch := make(chan string, 1)
	if _, err := conn.StartTransientUnitContext(ctx, sliceName, "replace", props, ch); err != nil {
		return fmt.Errorf("start transient slice %s: %w", sliceName, err)
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
