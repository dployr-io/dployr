package system

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"dployr/internal/scripts"
	"dployr/pkg/core/system"
	"dployr/pkg/core/utils"
	"dployr/pkg/shared"
	"dployr/pkg/store"
	"dployr/version"
)

var startTime = time.Now()

var (
	currentModeMu sync.RWMutex
	currentMode   = system.ModeReady
)

type DefaultService struct {
	cfg     *shared.Config
	store   store.InstanceStore
	results store.TaskResultStore
}

func NewDefaultService(cfg *shared.Config, store store.InstanceStore, results store.TaskResultStore) *DefaultService {
	return &DefaultService{cfg: cfg, store: store, results: results}
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

func (s *DefaultService) GetTasks(ctx context.Context, status string) (system.TaskSummary, error) {
	if status == "" || status == "pending" {
		return system.TaskSummary{Count: currentPendingTasks()}, nil
	}

	switch status {
	case "completed":
		if s.results == nil {
			return system.TaskSummary{Count: 0}, nil
		}
		results, err := s.results.ListUnsent(ctx)
		if err != nil {
			return system.TaskSummary{}, err
		}
		return system.TaskSummary{Count: len(results)}, nil
	default:
		return system.TaskSummary{}, fmt.Errorf("unsupported status %q", status)
	}
}

// GetMode returns the current daemon mode.
func (s *DefaultService) GetMode(ctx context.Context) (system.ModeStatus, error) {
	currentModeMu.RLock()
	defer currentModeMu.RUnlock()
	return system.ModeStatus{Mode: currentMode}, nil
}

// SetMode updates the daemon mode.
func (s *DefaultService) SetMode(ctx context.Context, req system.SetModeRequest) (system.ModeStatus, error) {
	mode := req.Mode
	if mode == "" {
		mode = system.ModeReady
	}

	if mode != system.ModeReady && mode != system.ModeUpdating {
		return system.ModeStatus{}, fmt.Errorf("unsupported mode %q", mode)
	}

	currentModeMu.Lock()
	currentMode = mode
	currentModeMu.Unlock()

	return system.ModeStatus{Mode: mode}, nil
}

// During the installation process, this method is used to register the instance with the base,
// used for routing traffic to this instance instead of directly hitting it.
// This is to ensure HTTPS traffic is enforced on dployr instance.
// Please refer to the documentation at https://docs.dployr.dev/installation for more details.
func (s *DefaultService) RequestDomain(ctx context.Context, req system.RequestDomainRequest) (string, error) {
	if s.cfg.BaseURL == "" {
		return "", fmt.Errorf("base_url is not configured")
	}

	if req.Token == "" {
		return "", fmt.Errorf("token cannot be empty")
	}

	if inst, err := s.store.GetInstance(ctx); err != nil {
		return "", fmt.Errorf("failed to get instance: %w", err)
	} else if inst != nil && inst.InstanceID != "" {
		fmt.Printf("instance already provisioned on %s\n", inst.RegisteredAt.Format(time.RFC3339))
		return "", nil
	}

	if err := s.store.SetToken(ctx, req.Token); err != nil {
		return "", fmt.Errorf("failed to set token: %w", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/v1/domains", s.cfg.BaseURL),
		"application/json",
		strings.NewReader(fmt.Sprintf(`{"token": "%s"}`, req.Token)))
	if err != nil {
		return "", fmt.Errorf("failed to assign domain: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Success bool `json:"success"`
			Error   struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return "", fmt.Errorf("failed to assign domain: %s", resp.Status)
		}
		return "", fmt.Errorf("%s (code: %s)", errResp.Error.Message, errResp.Error.Code)
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			InstanceID string `json:"instanceId"`
			Domain     string `json:"domain"`
			Issuer     string `json:"issuer"`
			Audience   string `json:"audience"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if !body.Success {
		return "", fmt.Errorf("base returned unsuccessful response without error payload")
	}

	if strings.TrimSpace(body.Data.Domain) == "" {
		return "", fmt.Errorf("base returned empty domain")
	}

	inst := &store.Instance{
		InstanceID: body.Data.InstanceID,
		Issuer:     body.Data.Issuer,
		Audience:   body.Data.Audience,
	}

	if err := s.store.RegisterInstance(ctx, inst); err != nil {
		return "", fmt.Errorf("failed to register instance: %w", err)
	}

	return body.Data.Domain, nil
}

func (s *DefaultService) RegisterInstance(ctx context.Context, req system.RegisterInstanceRequest) error {
	if req.Claim == "" {
		return fmt.Errorf("claim cannot be empty")
	}

	token, err := s.store.GetToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	if !compareBase64(token, req.Claim) {
		return fmt.Errorf("token does not match")
	}

	if req.InstanceID == "" {
		return fmt.Errorf("instance_id cannot be empty")
	}

	inst := &store.Instance{
		InstanceID: req.InstanceID,
		Issuer:     req.Issuer,
		Audience:   req.Audience,
	}

	if err := s.store.RegisterInstance(ctx, inst); err != nil {
		return err
	}

	return nil
}

func compareBase64(a, b string) bool {
	if a == "" || b == "" {
		return false
	}

	aBytes, errA := base64.StdEncoding.DecodeString(a)
	bBytes, errB := base64.StdEncoding.DecodeString(b)
	if errA != nil || errB != nil {
		return false
	}

	if len(aBytes) != len(bBytes) {
		return false
	}

	return subtle.ConstantTimeCompare(aBytes, bBytes) == 1
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
