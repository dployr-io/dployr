// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

// UpdateV1_1 is the v1.1 status update schema
type UpdateV1_1 struct {
	Schema      string          `json:"schema"` // Always "v1.1"
	Sequence    uint64          `json:"sequence"`
	Epoch       string          `json:"epoch"`
	InstanceID  string          `json:"instance_id"`
	Timestamp   string          `json:"timestamp"` // RFC3339
	IsFullSync  bool            `json:"is_full_sync"`
	Agent       *AgentInfo      `json:"agent,omitempty"`
	Status      StatusInfo      `json:"status"`
	Health      HealthInfo      `json:"health"`
	Resources   ResourcesInfo   `json:"resources"`
	Workloads   *WorkloadsInfo  `json:"workloads,omitempty"`
	Proxy       ProxyInfo       `json:"proxy"`
	Processes   ProcessesInfo   `json:"processes"`
	Filesystem  *FilesystemInfo `json:"filesystem,omitempty"`
	Diagnostics DiagnosticsInfo `json:"diagnostics"`
}

// AgentInfo - static agent metadata
type AgentInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// StatusInfo - operational status
type StatusInfo struct {
	State         string `json:"state"` // "healthy" | "degraded" | "unhealthy"
	Mode          string `json:"mode"`  // "ready" | "updating" | "maintenance"
	UptimeSeconds int64  `json:"uptime_seconds"`
}

// HealthInfo - component health breakdown
type HealthInfo struct {
	Overall   string `json:"overall"` // "ok" | "degraded" | "down"
	Websocket string `json:"websocket"`
	Tasks     string `json:"tasks"`
	Proxy     string `json:"proxy"`
	Auth      string `json:"auth"` // Advisory only
}

// ResourcesInfo - system resource metrics
type ResourcesInfo struct {
	CPU    CPUInfo    `json:"cpu"`
	Memory MemoryInfo `json:"memory"`
	Swap   SwapInfo   `json:"swap"`
	Disks  []DiskInfo `json:"disks,omitempty"`
}

type CPUInfo struct {
	Count         int          `json:"count"`
	UserPercent   float64      `json:"user_percent"`
	SystemPercent float64      `json:"system_percent"`
	IdlePercent   float64      `json:"idle_percent"`
	IOWaitPercent float64      `json:"iowait_percent"`
	LoadAverage   *LoadAvgInfo `json:"load_average,omitempty"` // null on Windows
}

type LoadAvgInfo struct {
	OneMinute     float64 `json:"one_minute"`
	FiveMinute    float64 `json:"five_minute"`
	FifteenMinute float64 `json:"fifteen_minute"`
}

type MemoryInfo struct {
	TotalBytes       int64 `json:"total_bytes"`
	UsedBytes        int64 `json:"used_bytes"`
	FreeBytes        int64 `json:"free_bytes"`
	AvailableBytes   int64 `json:"available_bytes"`
	BufferCacheBytes int64 `json:"buffer_cache_bytes"`
}

type SwapInfo struct {
	TotalBytes     int64 `json:"total_bytes"`
	UsedBytes      int64 `json:"used_bytes"`
	FreeBytes      int64 `json:"free_bytes"`
	AvailableBytes int64 `json:"available_bytes"`
}

type DiskInfo struct {
	Filesystem     string `json:"filesystem"`
	MountPoint     string `json:"mount_point"`
	TotalBytes     int64  `json:"total_bytes"`
	UsedBytes      int64  `json:"used_bytes"`
	AvailableBytes int64  `json:"available_bytes"`
}

// WorkloadsInfo - deployments and services
type WorkloadsInfo struct {
	Deployments []DeploymentV1_1 `json:"deployments"`
	Services    []ServiceV1_1    `json:"services"`
}

type DeploymentV1_1 struct {
	ID           string            `json:"id"`
	UserID       *string           `json:"user_id,omitempty"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Status       string            `json:"status"` // "pending" | "in_progress" | "completed" | "failed"
	Source       string            `json:"source"` // "remote" | "image" | "local"
	Runtime      RuntimeInfo       `json:"runtime"`
	Remote       *RemoteInfo       `json:"remote,omitempty"`
	Port         int               `json:"port"`
	WorkingDir   *string           `json:"working_dir,omitempty"`
	StaticDir    *string           `json:"static_dir,omitempty"`
	Image        *string           `json:"image,omitempty"`
	RunCommand   *string           `json:"run_command,omitempty"`
	BuildCommand *string           `json:"build_command,omitempty"`
	EnvVars      map[string]string `json:"env_vars,omitempty"`
	Secrets      []SecretRef       `json:"secrets,omitempty"`
	CreatedAt    string            `json:"created_at"`
	UpdatedAt    string            `json:"updated_at"`
}

type RuntimeInfo struct {
	Type    string  `json:"type"` // "static"|"golang"|"php"|"python"|"nodejs"|"ruby"|"dotnet"|"java"|"docker"|"k3s"|"custom"
	Version *string `json:"version,omitempty"`
}

type RemoteInfo struct {
	URL        string `json:"url"`
	Branch     string `json:"branch"`
	CommitHash string `json:"commit_hash"`
}

type SecretRef struct {
	Key    string `json:"key"`
	Source string `json:"source"` // "local"
}

type ServiceV1_1 struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	Runtime        string            `json:"runtime"`
	RuntimeVersion *string           `json:"runtime_version,omitempty"`
	Port           int               `json:"port"`
	WorkingDir     *string           `json:"working_dir,omitempty"`
	StaticDir      *string           `json:"static_dir,omitempty"`
	Image          *string           `json:"image,omitempty"`
	RunCommand     *string           `json:"run_command,omitempty"`
	BuildCommand   *string           `json:"build_command,omitempty"`
	EnvVars        map[string]string `json:"env_vars"`
	Secrets        []SecretRef       `json:"secrets"`
	RemoteURL      *string           `json:"remote_url,omitempty"`
	Branch         *string           `json:"branch,omitempty"`
	CommitHash     *string           `json:"commit_hash,omitempty"`
	CreatedAt      string            `json:"created_at"`
	UpdatedAt      string            `json:"updated_at"`
}

// ProxyInfo - reverse proxy state
type ProxyInfo struct {
	Type       string           `json:"type"`   // "caddy" | "nginx" | "apache" | "traefik" | "custom"
	Status     string           `json:"status"` // "running" | "stopped" | "unknown"
	Version    *string          `json:"version,omitempty"`
	RouteCount int              `json:"route_count"`
	Routes     []ProxyRouteInfo `json:"routes,omitempty"`
}

type ProxyRouteInfo struct {
	Domain   string  `json:"domain"`
	Upstream string  `json:"upstream"`
	Template string  `json:"template"` // "static" | "reverse_proxy" | "php_fastcgi"
	Root     *string `json:"root,omitempty"`
	Status   string  `json:"status"` // "active" | "inactive" | "error"
}

// ProcessesInfo - process metrics
type ProcessesInfo struct {
	Summary ProcessSummary    `json:"summary"`
	List    []ProcessInfoV1_1 `json:"list,omitempty"`
}

type ProcessSummary struct {
	Total    int `json:"total"`
	Running  int `json:"running"`
	Sleeping int `json:"sleeping"`
	Stopped  int `json:"stopped"`
	Zombie   int `json:"zombie"`
}

type ProcessInfoV1_1 struct {
	PID                 int     `json:"pid"`
	User                string  `json:"user"`
	Priority            int     `json:"priority"`
	Nice                int     `json:"nice"`
	VirtualMemoryBytes  int64   `json:"virtual_memory_bytes"`
	ResidentMemoryBytes int64   `json:"resident_memory_bytes"`
	SharedMemoryBytes   int64   `json:"shared_memory_bytes"`
	State               string  `json:"state"` // "running" | "sleeping" | "stopped" | "zombie" | "idle"
	CPUPercent          float64 `json:"cpu_percent"`
	MemoryPercent       float64 `json:"memory_percent"`
	CPUTime             string  `json:"cpu_time"`
	Command             string  `json:"command"`
}

// FilesystemInfo - filesystem snapshot
type FilesystemInfo struct {
	GeneratedAt string       `json:"generated_at"`
	IsStale     bool         `json:"is_stale"`
	Roots       []FSNodeV1_1 `json:"roots"`
}

type FSNodeV1_1 struct {
	Path          string        `json:"path"`
	Name          string        `json:"name"`
	Type          string        `json:"type"` // "file" | "directory" | "symlink"
	SizeBytes     int64         `json:"size_bytes"`
	ModifiedAt    string        `json:"modified_at"`
	Permissions   FSPermissions `json:"permissions"`
	Children      []FSNodeV1_1  `json:"children,omitempty"`
	IsTruncated   bool          `json:"is_truncated,omitempty"`
	TotalChildren *int          `json:"total_children,omitempty"`
}

type FSPermissions struct {
	Mode       string `json:"mode"`
	Owner      string `json:"owner"`
	Group      string `json:"group"`
	UID        int    `json:"uid"`
	GID        int    `json:"gid"`
	Readable   bool   `json:"readable"`
	Writable   bool   `json:"writable"`
	Executable bool   `json:"executable"`
}

// DiagnosticsInfo - debugging/observability
type DiagnosticsInfo struct {
	Websocket WebsocketDiag `json:"websocket"`
	Tasks     TasksDiag     `json:"tasks"`
	Auth      AuthDiag      `json:"auth"`
	Worker    *WorkerDiag   `json:"worker,omitempty"`
	Cert      *CertDiag     `json:"cert,omitempty"`
}

type WebsocketDiag struct {
	IsConnected     bool    `json:"is_connected"`
	LastConnectedAt *string `json:"last_connected_at,omitempty"`
	ReconnectCount  uint64  `json:"reconnect_count"`
	LastError       *string `json:"last_error,omitempty"`
}

type TasksDiag struct {
	InflightCount      int     `json:"inflight_count"`
	UnsentCount        int     `json:"unsent_count"`
	LastTaskID         *string `json:"last_task_id,omitempty"`
	LastTaskStatus     *string `json:"last_task_status,omitempty"`
	LastTaskDurationMs *int64  `json:"last_task_duration_ms,omitempty"`
	LastTaskAt         *string `json:"last_task_at,omitempty"`
}

type AuthDiag struct {
	TokenAgeSeconds       int64  `json:"token_age_seconds"`
	TokenExpiresInSeconds int64  `json:"token_expires_in_seconds"`
	BootstrapTokenPreview string `json:"bootstrap_token_preview"`
}

type WorkerDiag struct {
	MaxConcurrent int `json:"max_concurrent"`
	ActiveJobs    int `json:"active_jobs"`
}

type CertDiag struct {
	NotAfter      string `json:"not_after"`
	DaysRemaining int    `json:"days_remaining"`
}
