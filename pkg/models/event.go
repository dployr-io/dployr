package models

import (
	"time"
	"encoding/json"
)

// Event types
const (
    ProjectConfigured = "project_configured"
    ProjectCreated = "project_created"
    DeploymentStarted = "deployment_started"
    BuildStarted = "build_started"
    BuildCompleted = "build_completed"
    DeploymentCompleted = "deployment_completed"
    ServerProvisioned = "server_provisioned"
    ServiceConfigured = "service_configured"
    SSLCertificateObtained = "ssl_certificate_obtained"
    LogEvent = "log_event"
    Action = "action"
)

// Log Levels
const (
    LogLevelInfo = "info"
    LogLevelWarn = "warn"
    LogLevelError = "error"
)

type Event struct {
    ID          string          `db:"id" json:"id"`
    Type        string          `db:"type" json:"type"`
    AggregateID string          `db:"aggregate_id" json:"aggregate_id"`
    UserID      string          `db:"user_id" json:"user_id"`
    Timestamp   time.Time       `db:"timestamp" json:"timestamp"`
    Data        json.RawMessage `db:"data" json:"data"`
    Metadata    json.RawMessage `db:"metadata" json:"metadata"`
    Version     int             `db:"version" json:"version"`
}

type Metadata struct {
    CLIVersion    string `json:"cli_version"`
    UserAgent     string `json:"user_agent"`
    SourceIP      string `json:"source_ip"`
    CorrelationID string `json:"correlation_id"`
}

type ActionData struct {
    Name string `json:"name"`
}

type LogData struct {
    Level string `json:"level"`
    Message string `json:"message"`
}

type ProjectCreatedData struct {
    ProjectID string `json:"project_id"`
    Name      string `json:"name"`
    GitRepo   string `json:"git_repo"`
    Framework string `json:"framework"`
}

type ProjectConfiguredData struct {
    ProjectID string            `json:"project_id"`
    Domain    string            `json:"domain"`
    Provider  string            `json:"provider"`
    Config    map[string]string `json:"config"`
}

type DeploymentStartedData struct {
    DeploymentID string `json:"deployment_id"`
    ProjectID    string `json:"project_id"`
    CommitSHA    string `json:"commit_sha"`
    Branch       string `json:"branch"`
    Environment  string `json:"environment"`
}

type BuildStartedData struct {
    DeploymentID string `json:"deployment_id"`
    BuildID      string `json:"build_id"`
}

type BuildCompletedData struct {
    DeploymentID string        `json:"deployment_id"`
    BuildID      string        `json:"build_id"`
    Success      bool          `json:"success"`
    Duration     time.Duration `json:"duration"`
    Artifacts    []string      `json:"artifacts"`
    BuildSize    int64         `json:"build_size_bytes"`
}

type DeploymentCompletedData struct {
    DeploymentID string        `json:"deployment_id"`
    Status       string        `json:"status"` // "success", "failed"
    URL          string        `json:"url,omitempty"`
    Error        string        `json:"error,omitempty"`
    Duration     time.Duration `json:"duration"`
}

type ServerProvisionedData struct {
    ServerID  string `json:"server_id"`
    Provider  string `json:"provider"`
    Region    string `json:"region"`
    Size      string `json:"size"`
    PublicIP  string `json:"public_ip"`
    PrivateIP string `json:"private_ip"`
}

type ServiceConfiguredData struct {
    ServerID    string            `json:"server_id"`
    ServiceType string            `json:"service_type"` // "nginx", "pm2"
    Config      map[string]string `json:"config"`
    ConfigPath  string            `json:"config_path"`
}

type SSLCertificateObtainedData struct {
    Domain    string    `json:"domain"`
    Issuer    string    `json:"issuer"`
    ExpiresAt time.Time `json:"expires_at"`
    Method    string    `json:"method"`
}
