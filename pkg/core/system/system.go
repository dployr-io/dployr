package system

import (
	"context"
	"time"

	"github.com/dployr-io/dployr/pkg/core/utils"
	"github.com/dployr-io/dployr/version"
)

// Health status
const (
	HealthOK       = "ok"
	HealthDegraded = "degraded"
	HealthDown     = "down"
)

// System status
const (
	SystemStatusHealthy   = "healthy"
	SystemStatusDegraded  = "degraded"
	SystemStatusUnhealthy = "unhealthy"
)

// Proxy status
const (
	ProxyStatusRunning = "running"
	ProxyStatusStopped = "stopped"
	ProxyStatusUnknown = "unknown"
)

type DoctorResult struct {
	Status string `json:"status"`
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

// SystemServicesStatus describes aggregate service state.
type SystemServicesStatus struct {
	Total   int `json:"total"`
	Running int `json:"running"`
	Stopped int `json:"stopped"`
}

// SystemProxyStatus describes proxy health and routing information.
type SystemProxyStatus struct {
	Status string `json:"status"`
	Routes int    `json:"routes"`
}

// SystemStatus describes high-level health information about the daemon.
type SystemStatus struct {
	Status   string               `json:"status"`
	Mode     Mode                 `json:"mode"`
	Uptime   string               `json:"uptime"`
	Services SystemServicesStatus `json:"services"`
	Proxy    SystemProxyStatus    `json:"proxy"`
	Health   SystemHealth         `json:"health"`
	Debug    *SystemDebug         `json:"debug,omitempty"`
}

type SystemHealth struct {
	Overall string `json:"overall"` // ok|degraded|down
	WS      string `json:"ws"`
	Tasks   string `json:"tasks"`
	Auth    string `json:"auth"`
}

type SystemDebug struct {
	WS     WSDebug               `json:"ws"`
	Tasks  TasksDebug            `json:"tasks"`
	Auth   *AuthDebug            `json:"auth,omitempty"`
	Cert   *CertDebug            `json:"cert,omitempty"`
	System *SystemResourcesDebug `json:"system,omitempty"`
}

// SystemResourcesDebug provides high-level system resource information for debugging.
type SystemResourcesDebug struct {
	CPUCount       int              `json:"cpu_count"`
	MemTotalBytes  int64            `json:"mem_total_bytes,omitempty"`
	MemUsedBytes   int64            `json:"mem_used_bytes,omitempty"`
	MemFreeBytes   int64            `json:"mem_free_bytes,omitempty"`
	SwapTotalBytes int64            `json:"swap_total_bytes,omitempty"`
	SwapUsedBytes  int64            `json:"swap_used_bytes,omitempty"`
	Disks          []DiskDebugEntry `json:"disks,omitempty"`
}

// DiskDebugEntry represents disk usage for a single filesystem/mountpoint.
type DiskDebugEntry struct {
	Filesystem     string `json:"filesystem"`
	Mountpoint     string `json:"mountpoint"`
	SizeBytes      int64  `json:"size_bytes,omitempty"`
	UsedBytes      int64  `json:"used_bytes,omitempty"`
	AvailableBytes int64  `json:"available_bytes,omitempty"`
}

type WSDebug struct {
	Connected            bool    `json:"connected"`
	LastConnectAtRFC3339 string  `json:"last_connect_at"`
	ReconnectsSinceStart uint64  `json:"reconnects_since_start"`
	LastError            *string `json:"last_error,omitempty"`
}

type TasksDebug struct {
	Inflight          int    `json:"inflight"`
	DoneUnsent        int    `json:"done_unsent"`
	LastTaskID        string `json:"last_task_id,omitempty"`
	LastTaskStatus    string `json:"last_task_status,omitempty"`
	LastTaskDurMs     int64  `json:"last_task_dur_ms,omitempty"`
	LastTaskAtRFC3339 string `json:"last_task_at,omitempty"`
}

type AuthDebug struct {
	AgentTokenAgeS      int64 `json:"agent_token_age_s"`
	AgentTokenExpiresIn int64 `json:"agent_token_expires_in_s"`
}

type CertDebug struct {
	NotAfterRFC3339 string `json:"not_after"`
	DaysRemaining   int    `json:"days_remaining"`
}

type RegisterInstanceRequest struct {
	Claim      string `json:"claim"`
	InstanceID string `json:"instance_id"`
	Issuer     string `json:"issuer"`
	Audience   string `json:"audience"`
}

type RequestDomainRequest struct {
	Token string `json:"token"`
}

type RequestDomainResponse struct {
	Domain string `json:"domain"`
}

type InstallRequest struct {
	Version string `json:"version"`
}

// TaskSummary represents a simple count of tasks for a given status.
type TaskSummary struct {
	Count int `json:"count"`
}

// UpdateV1 represents the status update payload sent from the agent to base.
// This struct defines the core schema relied upon by clients interacting with the dployr API.
type UpdateV1 struct {
	Schema     string            `json:"schema"`
	Seq        uint64            `json:"seq"`
	Epoch      string            `json:"epoch"`
	Full       bool              `json:"full"`
	InstanceID string            `json:"instance_id"`
	BuildInfo  version.BuildInfo `json:"build_info"`
	Platform   PlatformInfo      `json:"platform"`
	Status     *SystemStatus     `json:"status,omitempty"`
}

// PlatformInfo describes the runtime platform of the agent.
type PlatformInfo struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

// HelloV1 is sent by the agent when establishing a WebSocket connection.
type HelloV1 struct {
	Schema           string            `json:"schema"`
	InstanceID       string            `json:"instance_id"`
	BuildInfo        version.BuildInfo `json:"build_info"`
	Platform         PlatformInfo      `json:"platform"`
	Capabilities     []string          `json:"capabilities,omitempty"`
	SchemasSupported []string          `json:"schemas_supported,omitempty"`
}

// HelloAckV1 is sent by base to acknowledge the agent hello.
type HelloAckV1 struct {
	Schema          string    `json:"schema"`
	Accept          bool      `json:"accept"`
	Reason          string    `json:"reason,omitempty"`
	FeaturesEnabled []string  `json:"features_enabled,omitempty"`
	ServerTime      time.Time `json:"server_time"`
}

// Mode represents the current mode.
// "ready"   – normal operation, syncer active.
// "updating" – in the middle of an update/installation cycle.
type Mode string

const (
	ModeReady    Mode = "ready"
	ModeUpdating Mode = "updating"
)

// ModeStatus describes the current daemon mode.
type ModeStatus struct {
	Mode Mode `json:"mode"`
}

// SetModeRequest is used by the agent to change the daemon mode.
type SetModeRequest struct {
	Mode Mode `json:"mode"`
}

// UpdateBootstrapTokenRequest is used to bootstrap token
type UpdateBootstrapTokenRequest struct {
	Token string `json:"token"`
}

// RegistrationStatus reports whether the instance has been registered
type RegistrationStatus struct {
	Registered bool `json:"registered"`
}

// System defines an interface for system operations.
type System interface {
	// GetInfo returns system information.
	GetInfo(ctx context.Context) (utils.SystemInfo, error)
	// RunDoctor runs the system doctor script and returns its combined output.
	RunDoctor(ctx context.Context) (string, error)
	// Install installs dployr; if version is empty, the latest version is installed.
	Install(ctx context.Context, version string) (string, error)
	// SystemStatus returns high-level health information.
	SystemStatus(ctx context.Context) (SystemStatus, error)
	// RequestDomain requests and assigns a new random domain from base to the system.
	RequestDomain(ctx context.Context, req RequestDomainRequest) (string, error)
	// RegisterInstance registers the system with the base and assigns an instance id
	RegisterInstance(ctx context.Context, req RegisterInstanceRequest) error
	// GetTasks returns a summary of tasks for the given status (e.g. "pending", "completed").
	GetTasks(ctx context.Context, status string) (TaskSummary, error)
	// GetMode returns the current daemon mode.
	GetMode(ctx context.Context) (ModeStatus, error)
	// SetMode updates the daemon mode.
	SetMode(ctx context.Context, req SetModeRequest) (ModeStatus, error)
	// UpdateBootstrapToken updates the bootstrap token.
	UpdateBootstrapToken(ctx context.Context, req UpdateBootstrapTokenRequest) error
	// IsRegistered returns true if this daemon has been registered with base.
	IsRegistered(ctx context.Context) (RegistrationStatus, error)
}

type SystemManager struct{}

func NewSystemManager() *SystemManager { return &SystemManager{} }
