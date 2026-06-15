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
	"strconv"
	"strings"
	"sync"
	"time"

	cgroup2 "github.com/containerd/cgroups/v3/cgroup2"
	systemddbus "github.com/coreos/go-systemd/v22/dbus"
	"github.com/dployr-io/dployr/pkg/core/system"
	godbus "github.com/godbus/dbus/v5"
)

type cpuSample struct {
	usageUsec uint64
	at        time.Time
}

var (
	cpuSamplesMu sync.Mutex
	cpuSamples   = map[string]cpuSample{} // keyed by cluster ID
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

		cpuLimit := readCPUMaxMillicores(slicePath)
		cpuUsagePct := computeCPUPercent(id, stats.CPU.GetUsageUsec())

		// Inactive file cache is reclaimable under pressure, so it doesn't count
		workingSetBytes := stats.Memory.Usage
		if stats.Memory.InactiveFile < workingSetBytes {
			workingSetBytes -= stats.Memory.InactiveFile
		} else {
			workingSetBytes = 0
		}

		result[id] = &system.ClusterResourcesInfo{
			MemoryUsedBytes:    int64(workingSetBytes),
			MemoryLimitBytes:   limit,
			CPULimitMillicores: cpuLimit,
			CPUUsagePercent:    cpuUsagePct,
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

// computeCPUPercent returns the CPU usage as a percentage of one full core since
// the last call for this cluster ID. Returns 0 on the first call (no prior sample).
func computeCPUPercent(clusterID string, usageUsec uint64) float64 {
	now := time.Now()

	cpuSamplesMu.Lock()
	prev, ok := cpuSamples[clusterID]
	cpuSamples[clusterID] = cpuSample{usageUsec: usageUsec, at: now}
	cpuSamplesMu.Unlock()

	if !ok {
		return 0
	}
	elapsedUs := now.Sub(prev.at).Microseconds()
	if elapsedUs <= 0 || usageUsec < prev.usageUsec {
		return 0
	}
	deltaUs := usageUsec - prev.usageUsec
	return float64(deltaUs) / float64(elapsedUs) * 100
}

// readCPUMaxMillicores parses cpu.max from a cgroup v2 slice directory.
// Format: "<quota_us> <period_us>" or "max <period_us>" (unlimited).
// Returns 0 when unlimited or unreadable.
func readCPUMaxMillicores(slicePath string) int64 {
	raw, err := os.ReadFile(filepath.Join(slicePath, "cpu.max"))
	if err != nil {
		return 0
	}
	fields := strings.Fields(strings.TrimSpace(string(raw)))
	if len(fields) != 2 || fields[0] == "max" {
		return 0
	}
	quota, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil || quota <= 0 {
		return 0
	}
	period, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil || period <= 0 {
		return 0
	}
	// millicores = (quota / period) * 1000
	return quota * 1000 / period
}
