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

// Service defines an interface for system operations.
type Service interface {
	GetInfo(ctx context.Context) (utils.SystemInfo, error)
	// RunDoctor runs the system doctor script and returns its combined output.
	RunDoctor(ctx context.Context) (string, error)
	// Install installs dployr; if version is empty, the latest version is installed.
	// After installation it runs the system doctor.
	Install(ctx context.Context, version string) (string, error)
}

type DefaultService struct{}

func NewDefaultService() *DefaultService { return &DefaultService{} }
