// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"runtime"
	"time"

	pkgAuth "github.com/dployr-io/dployr/pkg/auth"
	"github.com/dployr-io/dployr/pkg/core/proxy"
	"github.com/dployr-io/dployr/pkg/core/system"
	"github.com/dployr-io/dployr/pkg/core/utils"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"
	"github.com/dployr-io/dployr/version"
)

// BuildUpdateV1_1 constructs the v1.1 status update
func BuildUpdateV1_1(
	ctx context.Context,
	cfg *shared.Config,
	seq uint64,
	epoch string,
	isFullSync bool,
	instStore store.InstanceStore,
	deployStore store.DeploymentStore,
	svcStore store.ServiceStore,
	proxyHandler proxy.HandleProxy,
	fs *FileSystem,
	topCollector *TopCollector,
	workerMaxConcurrent int,
	workerActiveJobs int,
) (*system.UpdateV1_1, error) {
	now := time.Now()

	update := &system.UpdateV1_1{
		Schema:     "v1.1",
		Sequence:   seq,
		Epoch:      epoch,
		InstanceID: cfg.InstanceID,
		Timestamp:  now.Format(time.RFC3339Nano),
		IsFullSync: isFullSync,
	}

	update.Status = buildStatus()
	update.Health = buildHealth(ctx, instStore, proxyHandler)
	update.Resources = buildResources(ctx, topCollector, isFullSync)
	update.Proxy = buildProxy(proxyHandler, isFullSync)
	update.Processes = buildProcesses(ctx, topCollector, isFullSync)
	update.Diagnostics = buildDiagnostics(ctx, instStore, isFullSync, workerMaxConcurrent, workerActiveJobs)

	if isFullSync {
		update.Agent = buildAgent()
		workloads, err := buildWorkloads(ctx, deployStore, svcStore)
		if err != nil {
			return nil, err
		}
		update.Workloads = workloads
		update.Filesystem = buildFilesystem(fs)
	}

	return update, nil
}

func buildStatus() system.StatusInfo {
	currentModeMu.RLock()
	mode := currentMode
	currentModeMu.RUnlock()

	uptime := int64(time.Since(startTime).Seconds())

	state := "healthy"
	if mode == system.ModeUpdating {
		state = "degraded"
	}

	return system.StatusInfo{
		State:         state,
		Mode:          string(mode),
		UptimeSeconds: uptime,
	}
}

func buildHealth(ctx context.Context, instStore store.InstanceStore, proxyHandler proxy.HandleProxy) system.HealthInfo {
	wsHealth := system.HealthDown
	if WSConnected() {
		wsHealth = system.HealthOK
	}

	inflight := currentPendingTasks()
	tasksHealth := system.HealthOK
	if inflight > 100 {
		tasksHealth = system.HealthDegraded
	}

	authHealth, _ := computeAuthHealth(ctx, instStore)

	proxyHealth := system.HealthDown
	if proxyHandler != nil {
		proxyStatus := proxyHandler.Status()
		switch proxyStatus.Status {
		case "running":
			proxyHealth = system.HealthOK
		case "unknown":
			proxyHealth = system.HealthDegraded
		}
	}

	overall := worstHealth(wsHealth, tasksHealth, proxyHealth)

	return system.HealthInfo{
		Overall:   overall,
		Websocket: wsHealth,
		Tasks:     tasksHealth,
		Proxy:     proxyHealth,
		Auth:      authHealth,
	}
}

func buildResources(ctx context.Context, topCollector *TopCollector, includeDisks bool) system.ResourcesInfo {
	res := system.ResourcesInfo{
		CPU: system.CPUInfo{
			Count: runtime.NumCPU(),
		},
	}

	if topCollector == nil {
		return res
	}

	top, err := topCollector.Collect(ctx, "cpu", 10)
	if err != nil || top == nil {
		return res
	}

	res.CPU.UserPercent = top.CPU.User
	res.CPU.SystemPercent = top.CPU.System
	res.CPU.IdlePercent = top.CPU.Idle
	res.CPU.IOWaitPercent = top.CPU.Wait

	if top.Header.LoadAvg.One > 0 || top.Header.LoadAvg.Five > 0 {
		res.CPU.LoadAverage = &system.LoadAvgInfo{
			OneMinute:     top.Header.LoadAvg.One,
			FiveMinute:    top.Header.LoadAvg.Five,
			FifteenMinute: top.Header.LoadAvg.Fifteen,
		}
	}

	// Memory - convert MiB to bytes
	res.Memory = system.MemoryInfo{
		TotalBytes:       int64(top.Memory.Total * 1024 * 1024),
		UsedBytes:        int64(top.Memory.Used * 1024 * 1024),
		FreeBytes:        int64(top.Memory.Free * 1024 * 1024),
		AvailableBytes:   int64((top.Memory.Free + top.Memory.BufferCache) * 1024 * 1024),
		BufferCacheBytes: int64(top.Memory.BufferCache * 1024 * 1024),
	}

	// Swap - convert MiB to bytes
	res.Swap = system.SwapInfo{
		TotalBytes:     int64(top.Swap.Total * 1024 * 1024),
		UsedBytes:      int64(top.Swap.Used * 1024 * 1024),
		FreeBytes:      int64(top.Swap.Free * 1024 * 1024),
		AvailableBytes: int64(top.Swap.Available * 1024 * 1024),
	}

	if includeDisks {
		res.Disks = buildDisks()
	}

	return res
}

func buildDisks() []system.DiskInfo {
	sysInfo, err := utils.GetSystemInfo()
	if err != nil {
		return nil
	}

	var disks []system.DiskInfo
	for _, p := range sysInfo.Storage.Partitions {
		disks = append(disks, system.DiskInfo{
			Filesystem:     p.Filesystem,
			MountPoint:     p.Mountpoint,
			TotalBytes:     parseHumanBytes(p.Size),
			UsedBytes:      parseHumanBytes(p.Used),
			AvailableBytes: parseHumanBytes(p.Available),
		})
	}
	return disks
}

func buildProxy(proxyHandler proxy.HandleProxy, includeRoutes bool) system.ProxyInfo {
	info := system.ProxyInfo{
		Type:       "caddy",
		Status:     system.ProxyStatusUnknown,
		RouteCount: 0,
	}

	if proxyHandler == nil {
		return info
	}

	status := proxyHandler.Status()
	apps := proxyHandler.GetApps()

	info.Status = string(status.Status)
	info.RouteCount = len(apps)

	if status.Version != "" {
		info.Version = &status.Version
	}

	if includeRoutes {
		for _, app := range apps {
			route := system.ProxyRouteInfo{
				Domain:   app.Domain,
				Upstream: app.Upstream,
				Template: string(app.Template),
				Status:   "active",
			}
			if app.Root != "" {
				route.Root = &app.Root
			}
			info.Routes = append(info.Routes, route)
		}
	}

	return info
}

func buildProcesses(ctx context.Context, topCollector *TopCollector, includeList bool) system.ProcessesInfo {
	info := system.ProcessesInfo{
		Summary: system.ProcessSummary{},
	}

	if topCollector == nil {
		return info
	}

	top, err := topCollector.Collect(ctx, "cpu", 50)
	if err != nil || top == nil {
		return info
	}

	info.Summary = system.ProcessSummary{
		Total:    top.Tasks.Total,
		Running:  top.Tasks.Running,
		Sleeping: top.Tasks.Sleeping,
		Stopped:  top.Tasks.Stopped,
		Zombie:   top.Tasks.Zombie,
	}

	if includeList {
		for _, p := range top.Processes {
			info.List = append(info.List, system.ProcessInfoV1_1{
				PID:                 p.PID,
				User:                p.User,
				Priority:            p.Priority,
				Nice:                p.Nice,
				VirtualMemoryBytes:  p.VirtMem,
				ResidentMemoryBytes: p.ResMem,
				SharedMemoryBytes:   p.ShrMem,
				State:               mapProcessState(p.State),
				CPUPercent:          p.CPUPct,
				MemoryPercent:       p.MEMPct,
				CPUTime:             p.Time,
				Command:             p.Command,
			})
		}
	}

	return info
}

func mapProcessState(state string) string {
	switch state {
	case "S":
		return "sleeping"
	case "R":
		return "running"
	case "T":
		return "stopped"
	case "Z":
		return "zombie"
	case "I":
		return "idle"
	default:
		return state
	}
}

func buildDiagnostics(ctx context.Context, instStore store.InstanceStore, isFull bool, maxConcurrent, activeJobs int) system.DiagnosticsInfo {
	diag := system.DiagnosticsInfo{
		Websocket: buildWebsocketDiag(),
		Tasks:     buildTasksDiag(),
		Auth:      buildAuthDiag(ctx, instStore),
	}

	if isFull {
		diag.Worker = &system.WorkerDiag{
			MaxConcurrent: maxConcurrent,
			ActiveJobs:    activeJobs,
		}
		diag.Cert = buildCertDiag()
	}

	return diag
}

func buildWebsocketDiag() system.WebsocketDiag {
	diag := system.WebsocketDiag{
		IsConnected:    WSConnected(),
		ReconnectCount: WSReconnectsSinceStart(),
	}

	if t := WSLastConnect(); !t.IsZero() {
		s := t.Format(time.RFC3339)
		diag.LastConnectedAt = &s
	}

	if e := WSLastError(); e != nil {
		diag.LastError = e
	}

	return diag
}

func buildTasksDiag() system.TasksDiag {
	diag := system.TasksDiag{
		InflightCount: currentPendingTasks(),
		UnsentCount:   0,
	}

	if info := getLastExec(); info != nil {
		diag.LastTaskID = &info.ID
		diag.LastTaskStatus = &info.Status
		diag.LastTaskDurationMs = &info.DurMs
		t := info.At.Format(time.RFC3339)
		diag.LastTaskAt = &t
	}

	return diag
}

func buildAuthDiag(ctx context.Context, instStore store.InstanceStore) system.AuthDiag {
	_, debug := computeAuthHealth(ctx, instStore)

	diag := system.AuthDiag{}
	if debug != nil {
		diag.TokenAgeSeconds = debug.AgentTokenAgeS
		diag.TokenExpiresInSeconds = debug.AgentTokenExpiresIn
		diag.BootstrapTokenPreview = debug.BootstrapToken
	}
	return diag
}

func buildCertDiag() *system.CertDiag {
	certPath, _ := pkgAuth.DefaultClientCertPaths()

	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil
	}

	notAfter := cert.NotAfter.Format(time.RFC3339)
	daysRemaining := int(time.Until(cert.NotAfter).Hours() / 24)

	return &system.CertDiag{
		NotAfter:      notAfter,
		DaysRemaining: daysRemaining,
	}
}

func buildAgent() *system.AgentInfo {
	info := version.GetBuildInfo()
	return &system.AgentInfo{
		Version:   info.Version,
		Commit:    info.Commit,
		BuildDate: info.Date,
		GoVersion: info.GoVersion,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

func buildWorkloads(ctx context.Context, deployStore store.DeploymentStore, svcStore store.ServiceStore) (*system.WorkloadsInfo, error) {
	workloads := &system.WorkloadsInfo{
		Deployments: []system.DeploymentV1_1{},
		Services:    []system.ServiceV1_1{},
	}

	// Deployments
	if deployStore != nil {
		deps, err := deployStore.ListDeployments(ctx, 100, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to list deployments: %w", err)
		}
		if converted := system.FromStoreDeployments(deps); converted != nil {
			workloads.Deployments = converted
		}
	}

	// Services
	if svcStore != nil {
		svcs, err := svcStore.ListServices(ctx, 100, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to list services: %w", err)
		}
		if converted := system.FromStoreServices(svcs); converted != nil {
			workloads.Services = converted
		}
	}

	return workloads, nil
}

func buildFilesystem(fs *FileSystem) *system.FilesystemInfo {
	if fs == nil {
		return nil
	}

	snapshot := fs.GetSnapshot()
	if snapshot == nil {
		return nil
	}

	info := &system.FilesystemInfo{
		GeneratedAt: snapshot.GeneratedAt.Format(time.RFC3339Nano),
		IsStale:     snapshot.Stale,
	}

	for _, root := range snapshot.Roots {
		info.Roots = append(info.Roots, convertFSNode(root))
	}

	return info
}

func convertFSNode(n system.FSNode) system.FSNodeV1_1 {
	node := system.FSNodeV1_1{
		Path:       n.Path,
		Name:       n.Name,
		Type:       mapFSType(n.Type),
		SizeBytes:  n.SizeBytes,
		ModifiedAt: n.ModTime.Format(time.RFC3339),
		Permissions: system.FSPermissions{
			Mode:       n.Mode,
			Owner:      n.Owner,
			Group:      n.Group,
			UID:        n.UID,
			GID:        n.GID,
			Readable:   n.Readable,
			Writable:   n.Writable,
			Executable: n.Executable,
		},
		IsTruncated: n.Truncated,
	}

	if n.ChildCount > 0 {
		node.TotalChildren = &n.ChildCount
	}

	for _, child := range n.Children {
		node.Children = append(node.Children, convertFSNode(child))
	}

	return node
}

func mapFSType(t string) string {
	switch t {
	case "dir":
		return "directory"
	default:
		return t
	}
}
