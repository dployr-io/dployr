package system

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"dployr/internal/scripts"
	"dployr/pkg/core/utils"
	"dployr/version"
)

// Service defines a stable interface for system operations.
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

func (s *DefaultService) GetInfo(ctx context.Context) (utils.SystemInfo, error) {
	return utils.GetSystemInfo()
}

func (s *DefaultService) RunDoctor(ctx context.Context) (string, error) {
	ver := version.GetVersion()
	return runScriptWithEnv(ctx, scripts.SystemDoctorScript, []string{"DPLOYR_VERSION=" + ver})
}

func (s *DefaultService) Install(ctx context.Context, ver string) (string, error) {
	var env []string
	if ver != "" {
		env = append(env, "DPLOYR_VERSION="+ver)
	}

	out, err := runScriptWithEnv(ctx, scripts.SystemDoctorScript, env)
	if err != nil {
		return out, err
	}

	post, derr := s.RunDoctor(ctx)
	return out + post, derr
}

func runScriptWithEnv(ctx context.Context, script string, extraEnv []string) (string, error) {
	f, err := os.CreateTemp("", "system_doctor*.sh")
	if err != nil {
		return "", fmt.Errorf("failed to create temp script: %w", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(script); err != nil {
		f.Close()
		return "", fmt.Errorf("failed to write script: %w", err)
	}
	f.Close()

	if err := os.Chmod(f.Name(), 0o755); err != nil {
		return "", fmt.Errorf("failed to make script executable: %w", err)
	}

	cmd := exec.CommandContext(ctx, "bash", f.Name())
	cmd.Env = append(os.Environ(), extraEnv...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
