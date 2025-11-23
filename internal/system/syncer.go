package system

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"strings"
	"time"

	coresystem "dployr/pkg/core/system"
	"dployr/pkg/core/utils"
	"dployr/pkg/shared"
	"dployr/pkg/store"
	"dployr/pkg/tasks"
	"dployr/version"
)

type Syncer struct {
	cfg         *shared.Config
	logger      *slog.Logger
	instStore   store.InstanceStore
	resultStore store.TaskResultStore
	executor    *Executor
}

// syncRequest is the payload sent to base when polling for tasks.
type syncRequest struct {
	Version           string          `json:"version"`
	CompatibilityDate string          `json:"compatibility_date"`
	System            any             `json:"system"`
	CompletedTasks    []*tasks.Result `json:"completed_tasks"`
}

// syncTask represents a single task returned by base.
type syncTask struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Status  string          `json:"status"`
	Created int64           `json:"createdAt"`
	Updated int64           `json:"updatedAt"`
}

// syncResponse is the response envelope from base.
type syncResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Tasks []syncTask `json:"tasks"`
	} `json:"data"`
}

// agentTokenResponse is the response envelope from base when exchanging an
// instance credential for a short-lived agent access token.
type agentTokenResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

func NewSyncer(cfg *shared.Config, logger *slog.Logger, inst store.InstanceStore, results store.TaskResultStore, handler http.Handler) *Syncer {
	return &Syncer{
		cfg:         cfg,
		logger:      logger,
		instStore:   inst,
		resultStore: results,
		executor:    NewExecutor(logger, handler),
	}
}

func (s *Syncer) Start(ctx context.Context) {
	interval := shared.SanitizeSyncInterval(s.cfg.SyncInterval)
	if interval <= 0 {
		return
	}

	ticker := time.NewTicker(jitter(interval))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.syncOnce(ctx); err != nil {
				s.logger.Error("sync failed", "error", err)
			}
			ticker.Reset(jitter(interval))
		}
	}
}

func (s *Syncer) syncOnce(ctx context.Context) error {
	if s.cfg.BaseURL == "" {
		return fmt.Errorf("base_url is not configured")
	}

	// If the daemon is in updating mode, skip syncing until it returns to ready.
	currentModeMu.RLock()
	mode := currentMode
	currentModeMu.RUnlock()
	if mode == coresystem.ModeUpdating {
		s.logger.Info("skipping sync while daemon is in updating mode")
		return nil
	}

	inst, err := s.instStore.GetInstance(ctx)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}
	if inst == nil || strings.TrimSpace(inst.InstanceID) == "" {
		return nil
	}

	// Exchange the stored instance token for a short-lived agent token which
	// is then used to authenticate the status call.
	bootstrapToken, err := s.instStore.GetToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to load instance token: %w", err)
	}
	if strings.TrimSpace(bootstrapToken) == "" {
		// Instance has not been provisioned with a token yet; nothing to sync.
		return nil
	}

	agentToken, err := s.fetchAgentToken(ctx, bootstrapToken)
	if err != nil {
		return fmt.Errorf("failed to obtain agent token: %w", err)
	}

	sysInfo, err := utils.GetSystemInfo()
	if err != nil {
		s.logger.Warn("failed to get system info", "error", err)
	}

	// Load any previously completed tasks that have not yet been
	// acknowledged by base and include them in this request.
	pending, err := s.resultStore.ListUnsent(ctx)
	if err != nil {
		s.logger.Error("failed to load pending task results", "error", err)
	}

	body := syncRequest{
		Version:           version.GetVersion(),
		CompatibilityDate: time.Now().Format("2006-01-02"),
		System:            sysInfo,
		CompletedTasks:    toTaskResults(pending),
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal sync payload: %w", err)
	}

	url := fmt.Sprintf("%s/v1/agent/instances/%s/status", strings.TrimRight(s.cfg.BaseURL, "/"), inst.InstanceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to build sync request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+agentToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sync request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sync request returned status %d", resp.StatusCode)
	}

	var res syncResponse

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return fmt.Errorf("failed to decode sync response: %w", err)
	}

	if !res.Success {
		return fmt.Errorf("sync response marked unsuccessful: %s", res.Message)
	}

	// At this point base has accepted our completed task results,
	// so we can safely mark them as synced in the database.
	if len(pending) > 0 {
		ids := make([]string, 0, len(pending))
		for _, r := range pending {
			ids = append(ids, r.ID)
		}
		if err := s.resultStore.MarkSynced(ctx, ids); err != nil {
			s.logger.Error("failed to mark task results as synced", "error", err)
		}
	}

	if len(res.Data.Tasks) == 0 {
		return nil
	}

	s.logger.Info("received tasks from base", "count", len(res.Data.Tasks))

	// Execute tasks and collect results to be sent on the next sync.
	var completedTasks []*tasks.Result
	for _, t := range res.Data.Tasks {
		task := &tasks.Task{
			ID:      t.ID,
			Type:    t.Type,
			Payload: t.Payload,
			Status:  t.Status,
		}

		result := s.executor.Execute(ctx, task)
		completedTasks = append(completedTasks, result)
	}

	if err := s.resultStore.SaveResults(ctx, fromTaskResults(completedTasks)); err != nil {
		s.logger.Error("failed to persist completed task results", "error", err)
	}

	return nil
}

// fetchAgentToken exchanges a long-lived instance credential (bootstrap token)
// for a short-lived agent access token that can be used to authenticate agent
// calls (e.g. /v1/agent/instances/{id}/status).
func (s *Syncer) fetchAgentToken(ctx context.Context, bootstrapToken string) (string, error) {
	base := strings.TrimRight(s.cfg.BaseURL, "/")
	if base == "" {
		return "", fmt.Errorf("base_url is not configured")
	}

	url := fmt.Sprintf("%s/v1/agent/token", base)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to build agent token request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(bootstrapToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("agent token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("agent token request returned status %d", resp.StatusCode)
	}

	var body agentTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("failed to decode agent token response: %w", err)
	}
	if !body.Success || strings.TrimSpace(body.Data.Token) == "" {
		return "", fmt.Errorf("agent token response was unsuccessful or empty")
	}

	return body.Data.Token, nil
}

// toTaskResults converts stored TaskResult records into wire-level tasks.Result.
func toTaskResults(records []*store.TaskResult) []*tasks.Result {
	if len(records) == 0 {
		return nil
	}

	out := make([]*tasks.Result, 0, len(records))
	for _, r := range records {
		out = append(out, &tasks.Result{
			ID:     r.ID,
			Status: r.Status,
			Result: r.Result,
			Error:  r.Error,
		})
	}
	return out
}

// fromTaskResults converts wire-level tasks.Result into stored TaskResult
// records for persistence.
func fromTaskResults(results []*tasks.Result) []*store.TaskResult {
	if len(results) == 0 {
		return nil
	}

	out := make([]*store.TaskResult, 0, len(results))
	for _, r := range results {
		out = append(out, &store.TaskResult{
			ID:     r.ID,
			Status: r.Status,
			Result: r.Result,
			Error:  r.Error,
		})
	}
	return out
}

func jitter(base time.Duration) time.Duration {
	if base <= 0 {
		return base
	}

	delta := base / 10
	if delta <= 0 {
		return base
	}

	n := rand.Int63n(int64(2*delta+1)) - int64(delta)
	return base + time.Duration(n)
}
