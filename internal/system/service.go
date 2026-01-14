// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"bufio"
	"bytes"
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

	"github.com/dployr-io/dployr/internal/scripts"
	"github.com/dployr-io/dployr/pkg/core/system"
	"github.com/dployr-io/dployr/pkg/core/utils"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"
	"github.com/dployr-io/dployr/pkg/version"
	"github.com/golang-jwt/jwt/v4"
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
	ver := version.Short()
	return runSystemDoctorScript(ctx, scripts.SystemDoctorScript, []string{"DPLOYR_VERSION=" + ver}, s.store)
}

func (s *DefaultService) Install(ctx context.Context, req system.InstallRequest) (string, error) {
	version := req.Version
	if version == "" {
		version = "latest"
	}

	token := req.Token

	out, err := runInstallScript(ctx, scripts.InstallScript, version, token, s.store)
	if err != nil {
		return out, err
	}

	// Run doctor after installation to verify system health
	post, derr := s.RunDoctor(ctx)
	return out + "\n" + post, derr
}

func (s *DefaultService) Restart(ctx context.Context, req system.RestartRequest) (system.RestartResponse, error) {
	logger := shared.LogWithContext(ctx)

	// Check for pending tasks unless force is set
	if !req.Force {
		pending := currentPendingTasks()
		if pending > 0 {
			return system.RestartResponse{
				Status:  "rejected",
				Message: fmt.Sprintf("cannot restart: %d tasks are still running", pending),
			}, fmt.Errorf("cannot restart: %d tasks are still running", pending)
		}
	}

	logger.Info("initiating dployrd restart")

	// Set mode to updating to prevent new tasks
	currentModeMu.Lock()
	currentMode = system.ModeUpdating
	currentModeMu.Unlock()
	// Stop dployr first, then restart
	go func() {
		cmd := exec.Command("sudo", "systemctl", "restart", "dployrd")
		if err := cmd.Run(); err != nil {
			logger.Error("failed to restart dployrd", "error", err)
		}

		cmd = exec.Command("sudo", "systemctl", "restart", "caddy")
		if err := cmd.Run(); err != nil {
			logger.Error("failed to restart caddy", "error", err)
		}

		currentModeMu.Lock()
		currentMode = system.ModeReady
		currentModeMu.Unlock()
	}()

	return system.RestartResponse{
		Status:  "completed",
		Message: "dployrd restarted",
	}, nil
}

func (s *DefaultService) Reboot(ctx context.Context, req system.RebootRequest) (system.RebootResponse, error) {
	logger := shared.LogWithContext(ctx)

	if !req.Force {
		pending := currentPendingTasks()
		if pending > 0 {
			return system.RebootResponse{
				Status:  "rejected",
				Message: fmt.Sprintf("cannot reboot: %d tasks are still running", pending),
			}, fmt.Errorf("cannot reboot: %d tasks are still running", pending)
		}
	}

	logger.Info("initiating system reboot")

	currentModeMu.Lock()
	currentMode = system.ModeUpdating
	currentModeMu.Unlock()

	go func() {
		exec.Command("sudo", "systemctl", "stop", "dployrd").Run()
		time.Sleep(200 * time.Millisecond)

		cmd := exec.Command("sudo", "systemctl", "reboot")
		if err := cmd.Run(); err != nil {
			// Fallback
			exec.Command("sudo", "reboot").Run()
		}

		cmd = exec.Command("sudo", "systemctl", "restart", "caddy")
		if err := cmd.Run(); err != nil {
			logger.Error("failed to restart caddy", "error", err)
		}

		cmd = exec.Command("sudo", "systemctl", "start", "dployrd")
		if err := cmd.Run(); err != nil {
			logger.Error("failed to start dployrd", "error", err)
		}
		time.Sleep(200 * time.Millisecond)

		currentModeMu.Lock()
		currentMode = system.ModeReady
		currentModeMu.Unlock()
	}()

	return system.RebootResponse{
		Status:  "completed",
		Message: "system rebooted",
	}, nil
}

func (s *DefaultService) SystemStatus(ctx context.Context) (system.SystemStatus, error) {
	uptime := time.Since(startTime).Truncate(time.Second)

	var st system.SystemStatus
	st.Status = system.SystemStatusHealthy
	st.Mode = currentMode
	st.Uptime = uptime.String()
	st.Proxy.Status = system.ProxyStatusRunning
	st.Proxy.Routes = 0

	// Derive health (simple rules).
	wsOK := WSConnected()
	st.Health.WS = tern(wsOK, system.HealthOK, system.HealthDown)
	// Tasks
	inflight := currentPendingTasks()
	doneUnsent := 0
	if s.results != nil {
		if rs, err := s.results.ListUnsent(ctx); err == nil {
			doneUnsent = len(rs)
		}
	}
	if inflight == 0 && doneUnsent == 0 {
		st.Health.Tasks = system.HealthOK
	} else if !wsOK && inflight > 0 {
		st.Health.Tasks = system.HealthDegraded
	} else {
		st.Health.Tasks = system.HealthOK
	}

	// Auth health/debug derived from stored agent access token (JWT exp/iat).
	var authDbg *system.AuthDebug
	st.Health.Auth, authDbg = s.computeAuthHealthFromToken(ctx)
	st.Health.Overall = worst(st.Health.WS, st.Health.Tasks, st.Health.Auth)

	// Debug section
	dbg := system.SystemDebug{
		WS: system.WSDebug{
			Connected:            wsOK,
			LastConnectAtRFC3339: WSLastConnect().Format(time.RFC3339),
			ReconnectsSinceStart: WSReconnectsSinceStart(),
			LastError:            WSLastError(),
		},
		Tasks: system.TasksDebug{
			Inflight:   inflight,
			DoneUnsent: doneUnsent,
		},
	}
	if authDbg != nil {
		dbg.Auth = authDbg
	}
	if le := getLastExec(); le != nil {
		dbg.Tasks.LastTaskID = le.ID
		dbg.Tasks.LastTaskStatus = le.Status
		dbg.Tasks.LastTaskDurMs = le.DurMs
		dbg.Tasks.LastTaskAtRFC3339 = le.At.Format(time.RFC3339)
	}
	st.Debug = &dbg

	return st, nil
}

func tern[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

// computeAuthHealthFromToken checks the agent access token and returns health status and debug info.
func (s *DefaultService) computeAuthHealthFromToken(ctx context.Context) (string, *system.AuthDebug) {
	if s.store == nil {
		return system.HealthDown, nil
	}

	tok, err := s.store.GetBootstrapToken(ctx)
	if err != nil || strings.TrimSpace(tok) == "" {
		return system.HealthDown, nil
	}

	claims := &jwt.RegisteredClaims{}
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	if _, _, err := parser.ParseUnverified(strings.TrimSpace(tok), claims); err != nil {
		return system.HealthDown, nil
	}

	now := time.Now()
	var expTime, iatTime time.Time
	if claims.ExpiresAt != nil {
		expTime = claims.ExpiresAt.Time
	}
	if claims.IssuedAt != nil {
		iatTime = claims.IssuedAt.Time
	}

	age := int64(0)
	if !iatTime.IsZero() {
		age = int64(now.Sub(iatTime).Seconds())
	}
	ttl := int64(0)
	if !expTime.IsZero() {
		ttl = int64(expTime.Sub(now).Seconds())
	}
	if age < 0 {
		age = 0
	}
	if ttl < 0 {
		ttl = 0
	}

	authDbg := &system.AuthDebug{
		AgentTokenAgeS:      age,
		AgentTokenExpiresIn: ttl,
	}

	if ttl == 0 {
		return system.HealthDown, authDbg
	}
	return system.HealthOK, authDbg
}

// worst returns the most severe status string among the provided values.
func worst(vals ...string) string {
	rank := map[string]int{system.HealthOK: 0, system.HealthDegraded: 1, system.HealthDown: 2}
	w := system.HealthOK
	m := 0
	for _, v := range vals {
		if r, ok := rank[v]; ok && r > m {
			m = r
			w = v
		}
	}
	return w
}

func (s *DefaultService) GetTasks(ctx context.Context, status string, excludeSystem bool) (system.TaskSummary, error) {
	if status == "" || status == "pending" {
		var count int
		if excludeSystem {
			count = currentPendingTasksExcludingSystem()
		} else {
			count = currentPendingTasks()
		}
		return system.TaskSummary{Count: count}, nil
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

func (s *DefaultService) UpdateBootstrapToken(ctx context.Context, req system.UpdateBootstrapTokenRequest) error {
	if strings.TrimSpace(req.Token) == "" {
		return fmt.Errorf("bootstrap token cannot be empty")
	}

	return s.store.SetBootstrapToken(ctx, req.Token)
}

func (s *DefaultService) IsRegistered(ctx context.Context) (system.RegistrationStatus, error) {
	inst, err := s.store.GetInstance(ctx)
	if err != nil {
		return system.RegistrationStatus{}, err
	}
	if inst == nil {
		return system.RegistrationStatus{Registered: false}, nil
	}

	registered := strings.TrimSpace(inst.InstanceID) != "" && !inst.RegisteredAt.IsZero()
	if !registered {
		return system.RegistrationStatus{}, fmt.Errorf("instance registration incomplete; reinstall using a valid bootstrap token (see https://docs.dployr.io/installation)")
	}
	return system.RegistrationStatus{Registered: registered}, nil
}

// During the installation process, this method is used to register the instance with the base,
// used for routing traffic to this instance instead of directly hitting it.
// This is to ensure HTTPS traffic is enforced on dployr instance.
// Please refer to the documentation at https://docs.dployr.io/installation for more details.
func (s *DefaultService) RequestDomain(ctx context.Context, req system.RequestDomainRequest) (system.RequestDomainResponse, error) {
	if s.cfg.BaseURL == "" {
		return system.RequestDomainResponse{}, fmt.Errorf("base_url is not configured")
	}

	if req.Token == "" {
		return system.RequestDomainResponse{}, fmt.Errorf("token cannot be empty")
	}

	if inst, err := s.store.GetInstance(ctx); err != nil {
		return system.RequestDomainResponse{}, fmt.Errorf("failed to get instance: %w", err)
	} else if inst != nil && inst.InstanceID != "" {
		fmt.Printf("instance already provisioned on %s\n", inst.RegisteredAt.Format(time.RFC3339))
		return system.RequestDomainResponse{}, nil
	}

	if err := s.store.SetBootstrapToken(ctx, req.Token); err != nil {
		return system.RequestDomainResponse{}, fmt.Errorf("failed to set token: %w", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/v1/domains/register", s.cfg.BaseURL),
		"application/json",
		strings.NewReader(fmt.Sprintf(`{"token": "%s"}`, req.Token)))
	if err != nil {
		return system.RequestDomainResponse{}, fmt.Errorf("failed to assign domain: %w", err)
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
			return system.RequestDomainResponse{}, fmt.Errorf("failed to assign domain: %s", resp.Status)
		}
		return system.RequestDomainResponse{}, fmt.Errorf("%s (code: %s)", errResp.Error.Message, errResp.Error.Code)
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
		return system.RequestDomainResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if !body.Success {
		return system.RequestDomainResponse{}, fmt.Errorf("base returned unsuccessful response without error payload")
	}

	if strings.TrimSpace(body.Data.Domain) == "" {
		return system.RequestDomainResponse{}, fmt.Errorf("base returned empty domain")
	}

	inst := &store.Instance{
		InstanceID: body.Data.InstanceID,
		Issuer:     body.Data.Issuer,
		Audience:   body.Data.Audience,
	}

	if err := s.store.RegisterInstance(ctx, inst); err != nil {
		return system.RequestDomainResponse{}, fmt.Errorf("failed to register instance: %w", err)
	}

	return system.RequestDomainResponse{
		Success:    true,
		InstanceID: body.Data.InstanceID,
		Domain:     body.Data.Domain,
		Issuer:     body.Data.Issuer,
		Audience:   body.Data.Audience,
	}, nil
}

func (s *DefaultService) RegisterInstance(ctx context.Context, req system.RegisterInstanceRequest) error {
	if req.Claim == "" {
		return fmt.Errorf("claim cannot be empty")
	}

	token, err := s.store.GetBootstrapToken(ctx)
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

// streamCommandOutput runs a command and streams stdout/stderr line-by-line as structured logs.
// Returns the combined output and any error.
func streamCommandOutput(cmd *exec.Cmd, logger *shared.Logger) (string, error) {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	var output bytes.Buffer
	var wg sync.WaitGroup

	// Stream stdout as info logs
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			output.WriteString(line + "\n")
			logger.Info(line)
		}
	}()

	// Stream stderr as warn/error logs
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			output.WriteString(line + "\n")
			// Heuristic: lines with "error" or "fatal" are errors, others are warnings
			if strings.Contains(strings.ToLower(line), "error") || strings.Contains(strings.ToLower(line), "fatal") {
				logger.Error(line)
			} else {
				logger.Warn(line)
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Wait for streaming to complete
	wg.Wait()

	// Wait for command to finish
	err = cmd.Wait()

	return output.String(), err
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

func runInstallScript(ctx context.Context, script string, version string, token string, store store.InstanceStore) (string, error) {
	f, err := os.CreateTemp("", "install*.sh")
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

	// The embedded install script accepts: install.sh <version> [token]
	// Only pass the token argument if we actually have one.
	var cmd *exec.Cmd
	if strings.TrimSpace(token) != "" {
		cmd = exec.CommandContext(ctx, "bash", f.Name(), version, token)
	} else {
		cmd = exec.CommandContext(ctx, "bash", f.Name(), version)
	}
	cmd.Env = os.Environ()

	logger := shared.NewLogger().With("component", "installer", "script", "install", "version", version)
	out, err := streamCommandOutput(cmd, logger)
	if err == nil {
		store.UpdateLastInstalledAt(ctx)
	}
	return out, err
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

	logger := shared.NewLogger().With("component", "installer", "script", "system_doctor")
	out, err := streamCommandOutput(cmd, logger)
	store.UpdateLastInstalledAt(ctx)
	return out, err
}
