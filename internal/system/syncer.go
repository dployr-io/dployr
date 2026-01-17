// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/oklog/ulid/v2"

	pkgAuth "github.com/dployr-io/dployr/pkg/auth"
	"github.com/dployr-io/dployr/pkg/core/proxy"
	"github.com/dployr-io/dployr/pkg/core/system"
	"github.com/dployr-io/dployr/pkg/core/ws"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"
	"github.com/dployr-io/dployr/pkg/tasks"
	"github.com/dployr-io/dployr/version"
)

var wsConnected int32
var wsReconnects uint64
var lastWSConnect atomic.Value // time.Time
var lastWSError atomic.Value   // string

var wsConnectsTotal uint64
var wsDisconnectsTotal uint64
var agentTokenRefreshSuccessTotal uint64
var agentTokenRefreshFailedTotal uint64

var updateSeq uint64

func setWSConnected(v bool) {
	if v {
		atomic.StoreInt32(&wsConnected, 1)
	} else {
		atomic.StoreInt32(&wsConnected, 0)
	}
}

func computeAuthHealth(ctx context.Context, instStore store.InstanceStore) (health string, debug *system.AuthDebug) {
	return pkgAuth.ComputeAuthHealth(ctx, instStore)
}

func WSConnected() bool { return atomic.LoadInt32(&wsConnected) == 1 }

func WSReconnectsSinceStart() uint64 { return atomic.LoadUint64(&wsReconnects) }

func WSLastConnect() (t time.Time) {
	if v := lastWSConnect.Load(); v != nil {
		t = v.(time.Time)
	}
	return
}

func WSLastError() *string {
	if v := lastWSError.Load(); v != nil {
		s := v.(string)
		if s != "" {
			return &s
		}
	}
	return nil
}

func wsConnectTotal() uint64 { return atomic.LoadUint64(&wsConnectsTotal) }

func wsDisconnectTotal() uint64 { return atomic.LoadUint64(&wsDisconnectsTotal) }

func agentTokenRefreshTotals() (success, failed uint64) {
	return atomic.LoadUint64(&agentTokenRefreshSuccessTotal), atomic.LoadUint64(&agentTokenRefreshFailedTotal)
}

type syncerCtxKey string

const ctxKeyInstanceID syncerCtxKey = "instance_id"

const baseWSCACertPEM = ""

type Syncer struct {
	cfg                 *shared.Config
	logger              *shared.Logger
	instStore           store.InstanceStore
	resultStore         store.TaskResultStore
	deployStore         store.DeploymentStore
	svcStore            store.ServiceStore
	proxyHandler        proxy.HandleProxy
	fs                  *FileSystem
	topCollector        *TopCollector
	executor            *Executor
	agentTokenBackoff   time.Duration
	workerMaxConcurrent int
	workerActiveJobs    func() int
	epoch               string

	dedupeMu sync.Mutex
	dedupe   map[string]time.Time
}

func (s *Syncer) Executor() *Executor {
	return s.executor
}

func (s *Syncer) obtainAgentTokenWithBackoff(ctx context.Context, bootstrapToken string) (string, error) {
	return pkgAuth.ObtainAgentTokenWithBackoff(ctx, s.cfg.BaseURL, bootstrapToken, &s.agentTokenBackoff)
}

func (s *Syncer) ensureAccessToken(ctx context.Context, bootstrapToken string) (string, error) {
	accessToken, err := s.instStore.GetAccessToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load access token: %w", err)
	}
	if strings.TrimSpace(accessToken) != "" {
		s.logger.Debug("syncer: using cached access token")
		return accessToken, nil
	}

	s.logger.Info("syncer: obtaining agent access token")
	accessToken, err = s.obtainAgentTokenWithBackoff(ctx, bootstrapToken)
	if err != nil {
		atomic.AddUint64(&agentTokenRefreshFailedTotal, 1)
		return "", fmt.Errorf("failed to obtain agent token: %w", err)
	}
	atomic.AddUint64(&agentTokenRefreshSuccessTotal, 1)
	if err := s.instStore.SetAccessToken(ctx, accessToken); err != nil {
		s.logger.Error("syncer: failed to persist access token", "error", err)
	}
	s.logger.Debug("syncer: access token persisted")
	return accessToken, nil
}

func NewSyncer(cfg *shared.Config, logger *shared.Logger, instStore store.InstanceStore, resStore store.TaskResultStore, deployStore store.DeploymentStore, svcStore store.ServiceStore, proxyHandler proxy.HandleProxy, handler http.Handler, auth pkgAuth.Authenticator, fs *FileSystem, workerMaxConcurrent int, workerActiveJobs func() int) *Syncer {
	return &Syncer{
		cfg:                 cfg,
		logger:              logger,
		instStore:           instStore,
		resultStore:         resStore,
		deployStore:         deployStore,
		svcStore:            svcStore,
		proxyHandler:        proxyHandler,
		fs:                  fs,
		topCollector:        NewTopCollector(),
		executor:            NewExecutor(logger, cfg, handler, instStore, auth),
		workerMaxConcurrent: workerMaxConcurrent,
		workerActiveJobs:    workerActiveJobs,
		dedupe:              make(map[string]time.Time),
	}
}

func (s *Syncer) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := s.runWSConnection(ctx); err != nil {
				s.logger.Error("websocket connection failed", "error", err)
				time.Sleep(jitter(10 * time.Second))
			}
		}
	}
}

func (s *Syncer) runWSConnection(ctx context.Context) error {
	if s.cfg.BaseURL == "" {
		return fmt.Errorf("base_url is not configured")
	}

	s.logger.Debug("syncer: checking instance registration")
	inst, err := s.instStore.GetInstance(ctx)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}
	if inst == nil || strings.TrimSpace(inst.InstanceID) == "" {
		return fmt.Errorf("instance not registered; reinstall using a valid bootstrap token (see https://dployr.io/docs/quickstart.html)")
	}

	if s.epoch == "" {
		s.epoch = fmt.Sprintf("%s-%d", strings.TrimSpace(inst.InstanceID), startTime.Unix())
	}

	ctx = context.WithValue(ctx, ctxKeyInstanceID, inst.InstanceID)
	ctx = shared.EnrichContext(ctx)
	logger := s.logger.WithContext(ctx).With("instance_id", inst.InstanceID)

	logger.Debug("syncer: loading bootstrap token")
	bootstrapToken, err := s.instStore.GetBootstrapToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to load instance token: %w", err)
	}
	if strings.TrimSpace(bootstrapToken) == "" {
		logger.Debug("syncer: bootstrap token not set; skipping WS connect")
		return nil
	}

	accessToken, err := s.ensureAccessToken(ctx, bootstrapToken)
	if err != nil {
		return err
	}

	logger.Debug("syncer: ensuring client certificate", "instance_id", inst.InstanceID)
	clientCert, err := pkgAuth.EnsureClientCertificate(inst.InstanceID)
	if err != nil {
		return fmt.Errorf("failed to ensure client certificate: %w", err)
	}

	if len(clientCert.Certificate) > 0 {
		if parsed, err := x509.ParseCertificate(clientCert.Certificate[0]); err == nil {
			logger.Info("syncer: client certificate ready", "cn", parsed.Subject.CommonName, "not_after", parsed.NotAfter.Format(time.RFC3339))
		}
	}

	logger.Info("syncer: publishing client certificate")
	if err := pkgAuth.PublishClientCertificate(ctx, s.cfg.BaseURL, inst.InstanceID, accessToken, clientCert); err != nil {
		var httpErr *pkgAuth.HTTPError
		if errors.As(err, &httpErr) && (httpErr.StatusCode == http.StatusUnauthorized || httpErr.StatusCode == http.StatusForbidden) {
			logger.Warn("syncer: access token rejected; refreshing and retrying", "status", httpErr.StatusCode)
			if err := s.instStore.SetAccessToken(ctx, ""); err != nil {
				logger.Error("syncer: failed to clear invalid access token", "error", err)
			} else if accessToken, err = s.ensureAccessToken(ctx, bootstrapToken); err == nil {
				if err := pkgAuth.PublishClientCertificate(ctx, s.cfg.BaseURL, inst.InstanceID, accessToken, clientCert); err == nil {
					logger.Debug("syncer: cert publish succeeded after token refresh")
					goto ws_dial
				}
			}
		}
		return fmt.Errorf("failed to publish client certificate: %w", err)
	}
	logger.Debug("syncer: client certificate published")

ws_dial:

	wsURL := strings.Replace(s.cfg.BaseURL, "https://", "wss://", 1) +
		fmt.Sprintf("/v1/agent/ws?instanceName=%s", inst.InstanceID)

	logger.Info("syncer: dialing websocket", "host", s.cfg.BaseURL)
	tlsConfig, err := pkgAuth.BuildPinnedTLSConfig(clientCert, s.cfg.WSCertPath, baseWSCACertPEM)
	if err != nil {
		return fmt.Errorf("failed to build TLS config: %w", err)
	}
	logger.Debug("syncer: TLS config built", "pinned_roots", tlsConfig.RootCAs != nil)

	opts := &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Authorization": []string{"Bearer " + accessToken},
		},
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		},
	}

	conn, resp, err := websocket.Dial(ctx, wsURL, opts)
	if err != nil {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		logger.Debug("syncer: websocket dial failed", "status", status, "error", err)
		if status == http.StatusUnauthorized || status == http.StatusForbidden {
			logger.Warn("syncer: websocket upgrade rejected; refreshing token and retrying", "status", status)
			if err := s.instStore.SetAccessToken(ctx, ""); err != nil {
				logger.Error("syncer: failed to clear invalid access token", "error", err)
			} else if accessToken, err = s.ensureAccessToken(ctx, bootstrapToken); err == nil {
				opts.HTTPHeader.Set("Authorization", "Bearer "+accessToken)
				if conn2, resp2, err2 := websocket.Dial(ctx, wsURL, opts); err2 == nil {
					conn = conn2
					_ = resp2
					logger.Debug("syncer: websocket connected after token refresh")
					goto ws_connected
				}
			}
		}
		return fmt.Errorf("websocket dial failed: %w", err)
	}

ws_connected:

	defer conn.Close(websocket.StatusNormalClosure, "websocket connection closed")

	maxSize := s.cfg.WSMaxMessageSize
	if maxSize <= 0 {
		maxSize = 10 * 1024 * 1024 // 10MB default
	}
	conn.SetReadLimit(maxSize)

	logger.Info("syncer: websocket connected to base")
	atomic.AddUint64(&wsConnectsTotal, 1)
	setWSConnected(true)
	lastWSConnect.Store(time.Now())
	// clear last error on successful connect
	lastWSError.Store("")

	// Set WebSocket connection in executor for log streaming
	s.executor.SetWSConn(conn)
	defer s.executor.SetWSConn(nil)

	connCtx, cancelConn := context.WithCancel(ctx)
	defer cancelConn()

	{
		bi := version.GetBuildInfo()
		platform := system.PlatformInfo{OS: runtime.GOOS, Arch: runtime.GOARCH}
		h := &system.HelloV1{
			Schema:           "v1",
			InstanceID:       inst.InstanceID,
			BuildInfo:        bi,
			Platform:         platform,
			Capabilities:     []string{"tasks.v1", "updates.v1"},
			SchemasSupported: []string{"v1"},
		}
		_ = ws.Send(connCtx, conn, ws.Message{ID: ulid.Make().String(), TS: time.Now(), Kind: "hello", Hello: h})
	}

	pending, err := s.resultStore.ListUnsent(ctx)
	if err == nil && len(pending) > 0 {
		ids := make([]string, 0, len(pending))
		for _, r := range pending {
			ids = append(ids, r.ID)
		}
		logger.Debug("syncer: sending pending acks", "count", len(ids))
		if err := ws.Send(ctx, conn, ws.Message{ID: ulid.Make().String(), TS: time.Now(), Kind: "ack", IDs: ids}); err == nil {
			s.resultStore.MarkSynced(ctx, ids)
		}
	}

	logger.Debug("syncer: sending initial pull")
	if err := ws.Send(connCtx, conn, ws.Message{ID: ulid.Make().String(), TS: time.Now(), Kind: "pull"}); err != nil {
		return fmt.Errorf("failed to send initial pull: %w", err)
	}

	logger.Debug("syncer: sending immediate full sync on connect")
	seq := atomic.AddUint64(&updateSeq, 1)
	activeJobs := 0
	if s.workerActiveJobs != nil {
		activeJobs = s.workerActiveJobs()
	}
	update, err := BuildUpdateV1_1(
		connCtx,
		s.cfg,
		seq,
		s.epoch,
		true,
		s.instStore,
		s.deployStore,
		s.svcStore,
		s.proxyHandler,
		s.fs,
		s.topCollector,
		s.workerMaxConcurrent,
		activeJobs,
	)
	if err != nil {
		logger.Error("syncer: failed to build update", "error", err)
		return err
	}
	if err := ws.Send(connCtx, conn, ws.Message{
		ID:     ulid.Make().String(),
		TS:     time.Now(),
		Kind:   "update",
		Update: update,
	}); err != nil {
		logger.Error("syncer: failed to send immediate full sync", "error", err)
	}

	interval := shared.SanitizeSyncInterval(s.cfg.SyncInterval)
	go func() {
		for {
			select {
			case <-connCtx.Done():
				return
			case <-time.After(jitter(interval)):
				logger.Debug("syncer: sending periodic pull")
				if err := ws.Send(connCtx, conn, ws.Message{ID: ulid.Make().String(), TS: time.Now(), Kind: "pull"}); err != nil {
					logger.Error("syncer: failed to send periodic pull", "error", err)
					return
				}
			}
		}
	}()

	updateInterval := shared.SanitizeSyncInterval(s.cfg.SyncInterval)
	updateCounter := uint64(1)
	go func() {
		for {
			select {
			case <-connCtx.Done():
				return
			case <-time.After(jitter(updateInterval)):
				updateCounter++
				seq := atomic.AddUint64(&updateSeq, 1)

				activeJobs := 0
				if s.workerActiveJobs != nil {
					activeJobs = s.workerActiveJobs()
				}

				upd, err := BuildUpdateV1_1(
					connCtx,
					s.cfg,
					seq,
					s.epoch,
					true,
					s.instStore,
					s.deployStore,
					s.svcStore,
					s.proxyHandler,
					s.fs,
					s.topCollector,
					s.workerMaxConcurrent,
					activeJobs,
				)
				if err != nil {
					logger.Error("syncer: failed to build update", "error", err)
					continue
				}

				msg := ws.Message{
					ID:     ulid.Make().String(),
					TS:     time.Now(),
					Kind:   "update",
					Update: upd,
				}
				if err := ws.Send(connCtx, conn, msg); err != nil {
					logger.Error("syncer: failed to send update", "error", err)
					return
				}
			}
		}
	}()

	for {
		msg, err := ws.Read(connCtx, conn)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure || websocket.CloseStatus(err) == websocket.StatusGoingAway {
				logger.Info("syncer: websocket closed", "status", websocket.CloseStatus(err))
				atomic.AddUint64(&wsDisconnectsTotal, 1)
				setWSConnected(false)
				return nil
			}
			if errors.Is(err, io.EOF) {
				logger.Info("syncer: websocket closed (EOF)")
				atomic.AddUint64(&wsDisconnectsTotal, 1)
				setWSConnected(false)
				return nil
			}
			logger.Error("syncer: websocket read failed; will reconnect", "error", err)
			setWSConnected(false)
			atomic.AddUint64(&wsReconnects, 1)
			atomic.AddUint64(&wsDisconnectsTotal, 1)
			// cap error length to 256
			msgErr := err.Error()
			if len(msgErr) > 256 {
				msgErr = msgErr[:256]
			}
			lastWSError.Store(msgErr)
			return fmt.Errorf("websocket read failed: %w", err)
		}

		ctxMsg := shared.WithTrace(ctx, ulid.Make().String()) // new trace per WS message
		msgLogger := s.logger.WithContext(ctxMsg).With("instance_id", inst.InstanceID)

		msgLogger.Debug("syncer: received message", "kind", msg.Kind)
		switch msg.Kind {
		case "hello_ack":
			if msg.HelloAck != nil {
				if !msg.HelloAck.Accept {
					msgLogger.Warn("syncer: hello rejected by base", "reason", msg.HelloAck.Reason)
				}
				// features/hints can be used in future
				// inform about new minor, major version
			}
		case "task":
			msgLogger.Debug("syncer: received tasks", "count", len(msg.Items))
			s.handleTasks(ctxMsg, conn, msg.Items, msgLogger)
		}
	}
}

func (s *Syncer) handleTasks(ctx context.Context, conn *websocket.Conn, items []ws.Task, logger *shared.Logger) {
	currentModeMu.RLock()
	mode := currentMode
	currentModeMu.RUnlock()
	if mode == system.ModeUpdating {
		logger.Info("skipping tasks while daemon is in updating mode")
		return
	}

	if len(items) == 0 {
		return
	}

	logger.Info("received tasks from base", "count", len(items))

	var completedIDs []string
	var completedResults []*tasks.Result

	for _, t := range items {
		// Skip deduplication for log streaming tasks
		isLogStream := t.Type == "logs/stream:post"

		if !isLogStream && s.taskSeenRecently(t.ID) {
			logger.Info("skipping previously completed task", "task_id", t.ID)
			completedIDs = append(completedIDs, t.ID)
			continue
		}
		task := &tasks.Task{
			ID:      t.ID,
			Type:    t.Type,
			Payload: json.RawMessage(t.Payload),
			Status:  t.Status,
		}

		tctx := shared.WithRequest(ctx, t.ID)
		tctx = shared.WithTrace(tctx, ulid.Make().String())
		tlog := logger.WithContext(tctx)
		tlog.Debug("syncer: executing task", "task_id", t.ID, "type", t.Type)

		result := s.executor.Execute(tctx, task)

		result.Metadata = map[string]any{
			"instance_id": s.cfg.InstanceID,
		}

		// Send sync message when deploy task is completed
		if strings.Contains(t.Type, "deployments") || strings.Contains(t.Type, "proxy") {
			seq := atomic.AddUint64(&updateSeq, 1)
			activeJobs := 0
			if s.workerActiveJobs != nil {
				activeJobs = s.workerActiveJobs()
			}
			update, err := BuildUpdateV1_1(
				tctx,
				s.cfg,
				seq,
				s.epoch,
				true,
				s.instStore,
				s.deployStore,
				s.svcStore,
				s.proxyHandler,
				s.fs,
				s.topCollector,
				s.workerMaxConcurrent,
				activeJobs,
			)
			if err != nil {
				tlog.Error("syncer: failed to build update for deploy", "error", err)
			} else {
				if err := ws.Send(tctx, conn, ws.Message{
					ID:     ulid.Make().String(),
					TS:     time.Now(),
					Kind:   "update",
					Update: update,
				}); err != nil {
					tlog.Error("syncer: failed to send sync message for deploy", "error", err)
				}
			}
		}

		if err := s.sendTaskResponse(ctx, conn, t.ID, result); err != nil {
			tlog.Error("failed to send task_response", "error", err, "task_id", t.ID)
		} else {
			tlog.Debug("sent task_response", "task_id", t.ID, "success", result.Status == "done")
		}

		completedResults = append(completedResults, result)
		completedIDs = append(completedIDs, t.ID)

		if !isLogStream {
			s.markTaskSeen(t.ID)
		}
	}

	if err := s.resultStore.SaveResults(ctx, fromTaskResults(completedResults)); err != nil {
		logger.Error("failed to persist completed task results", "error", err)
		return
	}

	if len(completedIDs) > 0 {
		if err := ws.Send(ctx, conn, ws.Message{ID: ulid.Make().String(), TS: time.Now(), Kind: "ack", IDs: completedIDs}); err != nil {
			logger.Error("failed to send ack", "error", err)
		}
	}
}

// sendTaskResponse sends a task_response message back to the server for realtime requests
func (s *Syncer) sendTaskResponse(ctx context.Context, conn *websocket.Conn, taskID string, result *tasks.Result) error {
	success := result.Status == "done"

	msg := ws.Message{
		ID:        ulid.Make().String(),
		TS:        time.Now(),
		Kind:      "task_response",
		TaskID:    taskID,
		RequestID: taskID,
		Success:   success,
	}

	if success && result.Result != nil {
		msg.Data = result.Result
	} else if result.Metadata != nil {
		msg.Data = result.Metadata
	}

	if !success && result.Error != "" {
		msg.TaskError = &ws.TaskError{
			Code:    mapErrorToCode(result.Error),
			Message: result.Error,
		}
	}

	return ws.Send(ctx, conn, msg)
}

// mapErrorToCode maps error messages to WSErrorCode values
func mapErrorToCode(errMsg string) string {
	errLower := strings.ToLower(errMsg)
	switch {
	case strings.Contains(errLower, "permission denied"):
		return "PERMISSION_DENIED"
	case strings.Contains(errLower, "not found"):
		return "NOT_FOUND"
	case strings.Contains(errLower, "invalid"):
		return "INVALID_REQUEST"
	case strings.Contains(errLower, "timeout"):
		return "TIMEOUT"
	default:
		return "INTERNAL_ERROR"
	}
}

func (s *Syncer) taskSeenRecently(id string) bool {
	now := time.Now()
	s.dedupeMu.Lock()
	defer s.dedupeMu.Unlock()
	// cleanup expired entries opportunistically
	for k, exp := range s.dedupe {
		if now.After(exp) {
			delete(s.dedupe, k)
		}
	}
	if exp, ok := s.dedupe[id]; ok {
		if now.Before(exp) {
			return true
		}
		delete(s.dedupe, id)
	}
	return false
}

func (s *Syncer) markTaskSeen(id string) {
	ttl := s.cfg.TaskDedupTTL
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	s.dedupeMu.Lock()
	s.dedupe[id] = time.Now().Add(ttl)
	s.dedupeMu.Unlock()
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

// worstHealth returns the worst health status from a list of health statuses.
func worstHealth(vals ...string) string {
	rank := map[string]int{
		system.HealthOK:       0,
		system.HealthDegraded: 1,
		system.HealthDown:     2,
	}
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
