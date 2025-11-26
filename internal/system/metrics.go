package system

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dployr-io/dployr/pkg/core/utils"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"
	"github.com/dployr-io/dployr/version"
	"github.com/golang-jwt/jwt/v4"
)

type Metrics struct {
	cfg     *shared.Config
	inst    store.InstanceStore
	results store.TaskResultStore
}

func NewMetrics(cfg *shared.Config, inst store.InstanceStore, results store.TaskResultStore) *Metrics {
	return &Metrics{cfg: cfg, inst: inst, results: results}
}

func (m *Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	var buf bytes.Buffer

	bi := version.GetBuildInfo()
	commit := bi.Commit
	if len(commit) > 8 {
		commit = commit[:8]
	}
	buf.WriteString("# HELP dployr_build_info Build information\n")
	buf.WriteString("# TYPE dployr_build_info gauge\n")
	fmt.Fprintf(&buf, "dployr_build_info{version=\"%s\",commit=\"%s\",go_version=\"%s\"} 1\n\n", bi.Version, commit, bi.GoVersion)

	buf.WriteString("# HELP dployr_ws_connected Whether websocket to base is connected (1=yes)\n")
	buf.WriteString("# TYPE dployr_ws_connected gauge\n")
	wsConn := 0
	if WSConnected() {
		wsConn = 1
	}
	fmt.Fprintf(&buf, "dployr_ws_connected %d\n\n", wsConn)

	buf.WriteString("# HELP dployr_ws_connect_total Total successful websocket connections\n")
	buf.WriteString("# TYPE dployr_ws_connect_total counter\n")
	fmt.Fprintf(&buf, "dployr_ws_connect_total %d\n\n", wsConnectTotal())

	buf.WriteString("# HELP dployr_ws_disconnect_total Total websocket disconnects\n")
	buf.WriteString("# TYPE dployr_ws_disconnect_total counter\n")
	fmt.Fprintf(&buf, "dployr_ws_disconnect_total %d\n\n", wsDisconnectTotal())

	buf.WriteString("# HELP dployr_agent_token_expires_in_seconds Seconds until agent token expiry\n")
	buf.WriteString("# TYPE dployr_agent_token_expires_in_seconds gauge\n")
	fmt.Fprintf(&buf, "dployr_agent_token_expires_in_seconds %d\n\n", m.agentTokenTTLSeconds(ctx))

	buf.WriteString("# HELP dployr_tasks_inflight Tasks currently executing\n")
	buf.WriteString("# TYPE dployr_tasks_inflight gauge\n")
	fmt.Fprintf(&buf, "dployr_tasks_inflight %d\n\n", currentPendingTasks())

	buf.WriteString("# HELP dployr_tasks_done_unsent Completed task results not yet synced\n")
	buf.WriteString("# TYPE dployr_tasks_done_unsent gauge\n")
	fmt.Fprintf(&buf, "dployr_tasks_done_unsent %d\n\n", m.doneUnsent(ctx))

	buf.WriteString("# HELP dployr_task_executed_total Total tasks executed partitioned by result\n")
	buf.WriteString("# TYPE dployr_task_executed_total counter\n")
	ok, failed := taskExecutionTotals()
	fmt.Fprintf(&buf, "dployr_task_executed_total{result=\"success\"} %d\n", ok)
	fmt.Fprintf(&buf, "dployr_task_executed_total{result=\"failed\"} %d\n\n", failed)

	buf.WriteString("# HELP dployr_agent_token_refresh_total Total agent token refresh attempts partitioned by result\n")
	buf.WriteString("# TYPE dployr_agent_token_refresh_total counter\n")
	refOK, refFailed := agentTokenRefreshTotals()
	fmt.Fprintf(&buf, "dployr_agent_token_refresh_total{result=\"success\"} %d\n", refOK)
	fmt.Fprintf(&buf, "dployr_agent_token_refresh_total{result=\"failed\"} %d\n\n", refFailed)

	h := taskExecHistogramSnapshot()
	buf.WriteString("# HELP dployr_task_exec_seconds Task execution duration\n")
	buf.WriteString("# TYPE dployr_task_exec_seconds histogram\n")
	cum := uint64(0)
	for i, le := range h.Buckets {
		cum += h.Counts[i]
		leStr := strconv.FormatFloat(le, 'f', -1, 64)
		fmt.Fprintf(&buf, "dployr_task_exec_seconds_bucket{le=\"%s\"} %d\n", leStr, cum)
	}
	cum += h.Counts[len(h.Buckets)]
	fmt.Fprintf(&buf, "dployr_task_exec_seconds_bucket{le=\"+Inf\"} %d\n", cum)
	fmt.Fprintf(&buf, "dployr_task_exec_seconds_sum %f\n", h.Sum)
	fmt.Fprintf(&buf, "dployr_task_exec_seconds_count %d\n\n", h.Count)

	// System resource metrics
	m.writeSystemMetrics(&buf)

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write(buf.Bytes())
}

func (m *Metrics) doneUnsent(ctx context.Context) int {
	if m.results == nil {
		return 0
	}
	rs, err := m.results.ListUnsent(ctx)
	if err != nil {
		return 0
	}
	return len(rs)
}

func (m *Metrics) agentTokenTTLSeconds(ctx context.Context) int64 {
	if m.inst == nil {
		return 0
	}
	tok, err := m.inst.GetAccessToken(ctx)
	if err != nil {
		return 0
	}
	tok = strings.TrimSpace(tok)
	if tok == "" {
		return 0
	}
	claims := &jwt.RegisteredClaims{}
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	if _, _, err := parser.ParseUnverified(tok, claims); err != nil {
		return 0
	}
	if claims.ExpiresAt == nil {
		return 0
	}
	now := time.Now()
	ttl := int64(claims.ExpiresAt.Time.Sub(now).Seconds())
	if ttl < 0 {
		return 0
	}
	return ttl
}

func (m *Metrics) writeSystemMetrics(buf *bytes.Buffer) {
	sysInfo, err := utils.GetSystemInfo()
	if err != nil {
		return
	}

	// CPU count
	buf.WriteString("# HELP dployr_system_cpu_count Number of CPU cores\n")
	buf.WriteString("# TYPE dployr_system_cpu_count gauge\n")
	fmt.Fprintf(buf, "dployr_system_cpu_count %d\n\n", sysInfo.HW.CPUCount)

	// Memory metrics (parse human-readable values to bytes)
	if sysInfo.HW.MemTotal != nil {
		if bytes := parseHumanBytes(*sysInfo.HW.MemTotal); bytes > 0 {
			buf.WriteString("# HELP dployr_system_memory_total_bytes Total system memory in bytes\n")
			buf.WriteString("# TYPE dployr_system_memory_total_bytes gauge\n")
			fmt.Fprintf(buf, "dployr_system_memory_total_bytes %d\n\n", bytes)
		}
	}

	if sysInfo.HW.MemUsed != nil {
		if bytes := parseHumanBytes(*sysInfo.HW.MemUsed); bytes > 0 {
			buf.WriteString("# HELP dployr_system_memory_used_bytes Used system memory in bytes\n")
			buf.WriteString("# TYPE dployr_system_memory_used_bytes gauge\n")
			fmt.Fprintf(buf, "dployr_system_memory_used_bytes %d\n\n", bytes)
		}
	}

	if sysInfo.HW.MemFree != nil {
		if bytes := parseHumanBytes(*sysInfo.HW.MemFree); bytes > 0 {
			buf.WriteString("# HELP dployr_system_memory_free_bytes Free system memory in bytes\n")
			buf.WriteString("# TYPE dployr_system_memory_free_bytes gauge\n")
			fmt.Fprintf(buf, "dployr_system_memory_free_bytes %d\n\n", bytes)
		}
	}

	// Disk usage metrics for each partition
	for _, part := range sysInfo.Storage.Partitions {
		if size := parseHumanBytes(part.Size); size > 0 {
			buf.WriteString("# HELP dployr_system_disk_total_bytes Total disk space in bytes\n")
			buf.WriteString("# TYPE dployr_system_disk_total_bytes gauge\n")
			fmt.Fprintf(buf, "dployr_system_disk_total_bytes{mountpoint=\"%s\",filesystem=\"%s\"} %d\n", part.Mountpoint, part.Filesystem, size)
		}
		if used := parseHumanBytes(part.Used); used > 0 {
			buf.WriteString("# HELP dployr_system_disk_used_bytes Used disk space in bytes\n")
			buf.WriteString("# TYPE dployr_system_disk_used_bytes gauge\n")
			fmt.Fprintf(buf, "dployr_system_disk_used_bytes{mountpoint=\"%s\",filesystem=\"%s\"} %d\n", part.Mountpoint, part.Filesystem, used)
		}
		if avail := parseHumanBytes(part.Available); avail > 0 {
			buf.WriteString("# HELP dployr_system_disk_available_bytes Available disk space in bytes\n")
			buf.WriteString("# TYPE dployr_system_disk_available_bytes gauge\n")
			fmt.Fprintf(buf, "dployr_system_disk_available_bytes{mountpoint=\"%s\",filesystem=\"%s\"} %d\n\n", part.Mountpoint, part.Filesystem, avail)
		}
	}
}

// parseHumanBytes converts human-readable byte strings (e.g., "8.0G", "1.5T") to bytes.
// Returns 0 if parsing fails.
func parseHumanBytes(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return 0
	}

	// Extract numeric part and unit
	var val float64
	var unit string
	if _, err := fmt.Sscanf(s, "%f%s", &val, &unit); err != nil {
		return 0
	}

	// Convert based on unit (case-insensitive)
	unit = strings.ToUpper(unit)
	multiplier := int64(1)
	switch unit {
	case "K", "KB", "KI", "KIB":
		multiplier = 1024
	case "M", "MB", "MI", "MIB":
		multiplier = 1024 * 1024
	case "G", "GB", "GI", "GIB":
		multiplier = 1024 * 1024 * 1024
	case "T", "TB", "TI", "TIB":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "P", "PB", "PI", "PIB":
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024
	default:
		// No unit or unrecognized, assume bytes
		multiplier = 1
	}

	return int64(val * float64(multiplier))
}
