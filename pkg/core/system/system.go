package system

import (
	"context"
	"dployr/pkg/core/utils"
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
	Uptime   string               `json:"uptime"`
	Services SystemServicesStatus `json:"services"`
	Proxy    SystemProxyStatus    `json:"proxy"`
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
}

type SystemManager struct{}

func NewSystemManager() *SystemManager { return &SystemManager{} }
