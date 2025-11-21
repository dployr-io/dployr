package system

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"dployr/internal/scripts"
	"dployr/pkg/core/system"
	"dployr/pkg/core/utils"
	"dployr/pkg/shared"
	"dployr/pkg/store"
	"dployr/version"
)

var startTime = time.Now()

// Service defines an interface for system operations.
type Service interface {
	// GetInfo returns system information.
	GetInfo(ctx context.Context) (utils.SystemInfo, error)
	// RunDoctor runs the system doctor script and returns its combined output.
	RunDoctor(ctx context.Context) (string, error)
	// Install installs dployr; if version is empty, the latest version is installed.
	Install(ctx context.Context, version string) (string, error)
	// SystemStatus returns high-level health information.
	SystemStatus(ctx context.Context) (system.SystemStatus, error)
	// RequestDomain requests and assigns a new random domain from base to the system.
	RequestDomain(ctx context.Context, token string) error
	// RegisterInstance registers the system with the base and assigns an instance id
	RegisterInstance(token string) error
}

type DefaultService struct {
	cfg   *shared.Config
	store store.InstanceStore
}

func NewDefaultService(cfg *shared.Config, store store.InstanceStore) *DefaultService {
	return &DefaultService{cfg: cfg, store: store}
}

func (s *DefaultService) GetInfo(ctx context.Context) (utils.SystemInfo, error) {
	return utils.GetSystemInfo()
}

func (s *DefaultService) RunDoctor(ctx context.Context) (string, error) {
	ver := version.GetVersion()
	return runSystemDoctorScript(ctx, scripts.SystemDoctorScript, []string{"DPLOYR_VERSION=" + ver}, s.store)
}

func (s *DefaultService) Install(ctx context.Context, version string) (string, error) {
	var env []string
	if version != "" {
		env = append(env, "DPLOYR_VERSION="+version)
	}

	out, err := runSystemDoctorScript(ctx, scripts.SystemDoctorScript, env, s.store)
	if err != nil {
		return out, err
	}

	post, derr := s.RunDoctor(ctx)
	return out + post, derr
}

func (s *DefaultService) SystemStatus(ctx context.Context) (system.SystemStatus, error) {
	uptime := time.Since(startTime).Truncate(time.Second)

	var st system.SystemStatus
	st.Status = "healthy"
	st.Uptime = uptime.String()
	st.Services.Total = 0
	st.Services.Running = 0
	st.Services.Stopped = 0
	st.Proxy.Status = "running"
	st.Proxy.Routes = 0

	return st, nil
}

func (s *DefaultService) RequestDomain(ctx context.Context, token string) error {
	if s.cfg.BaseURL == "" {
		return fmt.Errorf("base_url is not configured")
	}

	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/v1/domains", s.cfg.BaseURL),
		"application/json",
		strings.NewReader(fmt.Sprintf(`{"token": "%s"}`, token)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to assign domain: %s", resp.Status)
	}

	return nil
}

func (s *DefaultService) RegisterInstance(ctx context.Context, req system.RegisterInstanceRequest) error {
	if req.Claim == "" {
		return fmt.Errorf("claim cannot be empty")
	}

	if token, err := s.store.GetToken(ctx); err != nil || token != req.Claim {
		return fmt.Errorf("token does not match")
	}

	if req.InstanceID == "" {
		return fmt.Errorf("instance_id cannot be empty")
	}

	i := &store.Instance{
		InstanceID: req.InstanceID,
		Token:      req.Claim,
		Issuer:     req.Issuer,
		Audience:   req.Audience,
	}

	if err := s.store.RegisterInstance(ctx, i); err != nil {
		return err
	}

	return nil
}

func runSystemDoctorScript(ctx context.Context, script string, extraEnv []string, store store.InstanceStore) (string, error) {
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
	store.UpdateLastInstalledAt(ctx)
	return string(out), err
}
