// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"context"
	"runtime"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"

	"github.com/dployr-io/dployr/pkg/core/system"
)

// TopCollector collects system resource usage data.
type TopCollector struct{}

// NewTopCollector creates a new TopCollector.
func NewTopCollector() *TopCollector {
	return &TopCollector{}
}

// Collect gathers a snapshot of system resource usage.
// sortBy can be "cpu" or "mem", limit is the max number of processes to return.
func (c *TopCollector) Collect(ctx context.Context, sortBy string, limit int) (*system.SystemTop, error) {
	if limit <= 0 {
		limit = 10
	}
	if sortBy == "" {
		sortBy = "cpu"
	}

	top := &system.SystemTop{}

	// Header - Time
	now := time.Now()
	top.Header.Time = now.Format("15:04:05")

	// Header - Uptime
	if uptime, err := host.UptimeWithContext(ctx); err == nil {
		top.Header.Uptime = time.Duration(uptime * 1000000000).String()
	}

	// Header - Users
	top.Header.Users = 0
	if users, err := host.Users(); err == nil {
		top.Header.Users = len(users)
	}

	// Header - Load average (not available on Windows)
	if runtime.GOOS != "windows" {
		if avg, err := load.AvgWithContext(ctx); err == nil {
			top.Header.LoadAvg = system.LoadAverage{
				One:     avg.Load1,
				Five:    avg.Load5,
				Fifteen: avg.Load15,
			}
		}
	}

	// CPU stats (aggregate across all cores)
	if cpuTimes, err := cpu.TimesWithContext(ctx, false); err == nil && len(cpuTimes) > 0 {
		t := cpuTimes[0]
		total := t.User + t.System + t.Idle + t.Iowait + t.Steal + t.Nice + t.Irq + t.Softirq
		if total > 0 {
			top.CPU = system.CPUStats{
				User:   (t.User / total) * 100,
				System: (t.System / total) * 100,
				Nice:   (t.Nice / total) * 100,
				Idle:   (t.Idle / total) * 100,
				Wait:   (t.Iowait / total) * 100,
				HI:     (t.Irq / total) * 100,
				SI:     (t.Softirq / total) * 100,
				ST:     (t.Steal / total) * 100,
			}
		}
	}

	// Memory stats (convert to MiB)
	if vmem, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		top.Memory = system.MemoryStats{
			Total:       float64(vmem.Total) / 1024 / 1024,
			Free:        float64(vmem.Free) / 1024 / 1024,
			Used:        float64(vmem.Used) / 1024 / 1024,
			BufferCache: float64(vmem.Buffers+vmem.Cached) / 1024 / 1024,
		}
	}

	// Swap stats (convert to MiB)
	if swap, err := mem.SwapMemoryWithContext(ctx); err == nil {
		top.Swap = system.SwapStats{
			Total:     float64(swap.Total) / 1024 / 1024,
			Free:      float64(swap.Free) / 1024 / 1024,
			Used:      float64(swap.Used) / 1024 / 1024,
			Available: float64(swap.Free) / 1024 / 1024,
		}
	}

	// Process list
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return top, nil // return what we have
	}

	var procInfos []system.ProcessInfo
	taskStats := system.TaskStats{Total: len(procs)}

	for _, p := range procs {
		// Get process state for task stats
		status, _ := p.StatusWithContext(ctx)
		if len(status) > 0 {
			switch status[0] {
			case process.Running:
				taskStats.Running++
			case process.Sleep:
				taskStats.Sleeping++
			case process.Stop:
				taskStats.Stopped++
			case process.Zombie:
				taskStats.Zombie++
			}
		}

		// Collect process info
		info := system.ProcessInfo{
			PID: int(p.Pid),
		}

		if name, err := p.NameWithContext(ctx); err == nil {
			info.Command = name
		}

		if user, err := p.UsernameWithContext(ctx); err == nil {
			info.User = user
		}

		// Get nice value and priority
		if nice, err := p.NiceWithContext(ctx); err == nil {
			info.Nice = int(nice)
		}
		info.Priority = 20 // Default priority

		if cpuPct, err := p.CPUPercentWithContext(ctx); err == nil {
			info.CPUPct = cpuPct
		}

		if memPct, err := p.MemoryPercentWithContext(ctx); err == nil {
			info.MEMPct = float64(memPct)
		}

		if memInfo, err := p.MemoryInfoWithContext(ctx); err == nil && memInfo != nil {
			info.ResMem = int64(memInfo.RSS)
			info.VirtMem = int64(memInfo.VMS)
			info.ShrMem = 0 // Shared memory not available in gopsutil
		}

		if len(status) > 0 {
			info.State = string(status[0])
		}

		// Get CPU time
		if times, err := p.TimesWithContext(ctx); err == nil {
			totalSecs := int(times.User + times.System)
			info.Time = time.Duration(totalSecs * 1000000000).String()
		}

		procInfos = append(procInfos, info)
	}

	top.Tasks = taskStats

	// Sort processes
	switch sortBy {
	case "mem":
		sort.Slice(procInfos, func(i, j int) bool {
			return procInfos[i].MEMPct > procInfos[j].MEMPct
		})
	default: // "cpu"
		sort.Slice(procInfos, func(i, j int) bool {
			return procInfos[i].CPUPct > procInfos[j].CPUPct
		})
	}

	// Limit results
	if len(procInfos) > limit {
		procInfos = procInfos[:limit]
	}

	top.Processes = procInfos

	return top, nil
}

// CollectSummary returns a lightweight summary suitable for WS updates.
// Returns top 10 processes by CPU.
func (c *TopCollector) CollectSummary(ctx context.Context) (*system.SystemTop, error) {
	return c.Collect(ctx, "cpu", 10)
}
