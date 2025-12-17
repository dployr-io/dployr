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

	top := &system.SystemTop{
		Timestamp: time.Now(),
	}

	// Uptime
	if uptime, err := host.UptimeWithContext(ctx); err == nil {
		top.Uptime = uptime
	}

	// Load average (not available on Windows)
	if runtime.GOOS != "windows" {
		if avg, err := load.AvgWithContext(ctx); err == nil {
			top.LoadAvg = system.LoadAverage{
				Load1:  avg.Load1,
				Load5:  avg.Load5,
				Load15: avg.Load15,
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
				Idle:   (t.Idle / total) * 100,
				IOWait: (t.Iowait / total) * 100,
				Steal:  (t.Steal / total) * 100,
			}
		}
	}

	// Memory stats
	if vmem, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		top.Memory = system.MemoryStats{
			Total:       vmem.Total,
			Used:        vmem.Used,
			Free:        vmem.Free,
			Available:   vmem.Available,
			UsedPercent: vmem.UsedPercent,
		}
	}

	// Swap stats
	if swap, err := mem.SwapMemoryWithContext(ctx); err == nil {
		top.Memory.SwapTotal = swap.Total
		top.Memory.SwapUsed = swap.Used
		top.Memory.SwapFree = swap.Free
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
			PID: p.Pid,
		}

		if name, err := p.NameWithContext(ctx); err == nil {
			info.Command = name
		}

		if user, err := p.UsernameWithContext(ctx); err == nil {
			info.User = user
		}

		if cpuPct, err := p.CPUPercentWithContext(ctx); err == nil {
			info.CPUPercent = cpuPct
		}

		if memPct, err := p.MemoryPercentWithContext(ctx); err == nil {
			info.MemPercent = memPct
		}

		if memInfo, err := p.MemoryInfoWithContext(ctx); err == nil && memInfo != nil {
			info.RSS = int64(memInfo.RSS)
			info.VMS = int64(memInfo.VMS)
		}

		if len(status) > 0 {
			info.State = string(status[0])
		}

		procInfos = append(procInfos, info)
	}

	top.Tasks = taskStats

	// Sort processes
	switch sortBy {
	case "mem":
		sort.Slice(procInfos, func(i, j int) bool {
			return procInfos[i].MemPercent > procInfos[j].MemPercent
		})
	default: // "cpu"
		sort.Slice(procInfos, func(i, j int) bool {
			return procInfos[i].CPUPercent > procInfos[j].CPUPercent
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
