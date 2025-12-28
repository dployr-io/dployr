// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/golang-jwt/jwt/v4"
	"github.com/oklog/ulid/v2"

	pkgAuth "github.com/dployr-io/dployr/pkg/auth"
	"github.com/dployr-io/dployr/pkg/core/logs"
	"github.com/dployr-io/dployr/pkg/core/proxy"
	"github.com/dployr-io/dployr/pkg/core/system"
	"github.com/dployr-io/dployr/pkg/core/utils"
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

// TaskError represents an error in task response
type TaskError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// wsMessage represents WebSocket messages exchanged with base.
type wsMessage struct {
	ID        string             `json:"id,omitempty"`
	RequestID string             `json:"request_id,omitempty"`
	TS        time.Time          `json:"ts"`
	Kind      string             `json:"kind"`
	Items     []syncTask         `json:"items,omitempty"`
	IDs       []string           `json:"ids,omitempty"`
	Update    *system.UpdateV1   `json:"update,omitempty"`
	Hello     *system.HelloV1    `json:"hello,omitempty"`
	HelloAck  *system.HelloAckV1 `json:"hello_ack,omitempty"`
	LogChunk  *logs.LogChunk     `json:"log_chunk,omitempty"`
	// Task response fields for filesystem operations
	TaskID    string      `json:"taskId,omitempty"`
	Success   bool        `json:"success,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	TaskError *TaskError  `json:"error,omitempty"`
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

// agentTokenResponse is the response envelope from base when exchanging an
// instance credential for a short-lived agent access token.
type agentTokenResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

// computeAuthHealth checks the agent access token and returns health status and debug info
func computeAuthHealth(ctx context.Context, instStore store.InstanceStore) (health string, debug *system.AuthDebug) {
	health = system.HealthDown
	if instStore == nil {
		return
	}

	tok, err := instStore.GetAccessToken(ctx)
	if err != nil || strings.TrimSpace(tok) == "" {
		return
	}

	bTok, err := instStore.GetBootstrapToken(ctx)
	if err != nil || strings.TrimSpace(bTok) == "" {
		return
	}

	const prevLen = 70
	if len(bTok) > prevLen {
		bTok = bTok[:prevLen]
	}

	claims := &jwt.RegisteredClaims{}
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	if _, _, err := parser.ParseUnverified(strings.TrimSpace(tok), claims); err != nil {
		return
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

	debug = &system.AuthDebug{
		AgentTokenAgeS:      age,
		AgentTokenExpiresIn: ttl,
		BootstrapToken:      bTok,
	}

	if ttl == 0 {
		health = system.HealthDown
	} else {
		health = system.HealthOK
	}

	return
}

func buildAgentUpdate(ctx context.Context, cfg *shared.Config, instanceID string, instStore store.InstanceStore, deployStore store.DeploymentStore, svcStore store.ServiceStore, proxyHandler proxy.HandleProxy, fs *FileSystem) *system.UpdateV1 {
	seq := atomic.AddUint64(&updateSeq, 1)
	uptime := time.Since(startTime).Seconds()
	currentModeMu.RLock()
	mode := currentMode
	currentModeMu.RUnlock()

	// Compute health status
	wsOK := WSConnected()
	wsHealth := system.HealthDown
	if wsOK {
		wsHealth = system.HealthOK
	}

	inflight := currentPendingTasks()
	tasksHealth := system.HealthOK
	if inflight > 0 && !wsOK {
		tasksHealth = system.HealthDegraded
	}

	// Auth health: check agent access token expiration
	authHealth, authDbg := computeAuthHealth(ctx, instStore)

	// Overall health: worst of WS, Tasks, Auth
	overallHealth := worstHealth(wsHealth, tasksHealth, authHealth)

	// Build debug info
	dbg := &system.SystemDebug{
		WS: system.WSDebug{
			Connected:            wsOK,
			LastConnectAtRFC3339: WSLastConnect().Format(time.RFC3339),
			ReconnectsSinceStart: WSReconnectsSinceStart(),
			LastError:            WSLastError(),
		},
		Tasks: system.TasksDebug{
			Inflight:   inflight,
			DoneUnsent: 0, // Not available in this context
		},
		Auth: authDbg,
	}
	if le := getLastExec(); le != nil {
		dbg.Tasks.LastTaskID = le.ID
		dbg.Tasks.LastTaskStatus = le.Status
		dbg.Tasks.LastTaskDurMs = le.DurMs
		dbg.Tasks.LastTaskAtRFC3339 = le.At.Format(time.RFC3339)
	}

	// Populate system resource details (CPU, memory, disk, workers) into debug struct.
	if sysInfo, err := utils.GetSystemInfo(); err == nil {
		res := &system.SystemResourcesDebug{
			CPUCount: sysInfo.HW.CPUCount,
			Workers:  cfg.MaxWorkers,
		}
		if sysInfo.HW.MemTotal != nil {
			if v := parseHumanBytes(*sysInfo.HW.MemTotal); v > 0 {
				res.MemTotalBytes = v
			}
		}
		if sysInfo.HW.MemUsed != nil {
			if v := parseHumanBytes(*sysInfo.HW.MemUsed); v > 0 {
				res.MemUsedBytes = v
			}
		}
		if sysInfo.HW.MemFree != nil {
			if v := parseHumanBytes(*sysInfo.HW.MemFree); v > 0 {
				res.MemFreeBytes = v
			}
		}
		for _, p := range sysInfo.Storage.Partitions {
			entry := system.DiskDebugEntry{
				Filesystem: p.Filesystem,
				Mountpoint: p.Mountpoint,
			}
			if v := parseHumanBytes(p.Size); v > 0 {
				entry.SizeBytes = v
			}
			if v := parseHumanBytes(p.Used); v > 0 {
				entry.UsedBytes = v
			}
			if v := parseHumanBytes(p.Available); v > 0 {
				entry.AvailableBytes = v
			}
			res.Disks = append(res.Disks, entry)
		}
		dbg.System = res
	}

	deployments := make([]store.Deployment, 0)
	if deployStore != nil {
		if deps, err := deployStore.ListDeployments(ctx, 100, 0); err == nil {
			for _, d := range deps {
				sanitized := *d
				if len(sanitized.Blueprint.Secrets) > 0 {
					keys := make(map[string]string, len(sanitized.Blueprint.Secrets))
					for key := range d.Blueprint.Secrets {
						keys[key] = ""
					}
					sanitized.Blueprint.Secrets = keys
				}
				deployments = append(deployments, sanitized)
			}
		}
	}

	services := make([]store.Service, 0)
	if svcStore != nil {
		if svcs, err := svcStore.ListServices(ctx, 100, 0); err == nil {
			for _, s := range svcs {
				sanitized := *s
				if len(sanitized.Secrets) > 0 {
					keys := make(map[string]string, len(sanitized.Secrets))
					for key := range s.Secrets {
						keys[key] = ""
					}
					sanitized.Secrets = keys
				}
				services = append(services, sanitized)
			}
		}
	}

	apps := make([]proxy.App, 0)
	if proxyHandler != nil {
		apps = proxyHandler.GetApps()
		if apps == nil {
			apps = make([]proxy.App, 0)
		}
	}

	var fsSnapshot *system.FSSnapshot
	if fs != nil {
		fsSnapshot = fs.GetSnapshot()
	}

	return &system.UpdateV1{
		Schema:      "v1",
		Seq:         seq,
		Epoch:       fmt.Sprintf("%s-%d", strings.TrimSpace(instanceID), startTime.Unix()),
		Full:        false,
		InstanceID:  cfg.InstanceID,
		BuildInfo:   version.GetBuildInfo(),
		Platform:    system.PlatformInfo{OS: runtime.GOOS, Arch: runtime.GOARCH},
		Status:      system.SystemStatusHealthy,
		Mode:        mode,
		Uptime:      strconv.FormatInt(int64(uptime), 10),
		Deployments: deployments,
		Services:    services,
		Apps:        apps,
		Proxy:       system.SystemProxyStatus{Status: system.ProxyStatusRunning, Routes: len(apps)},
		Health:      system.SystemHealth{Overall: overallHealth, WS: wsHealth, Tasks: tasksHealth, Auth: authHealth},
		Debug:       dbg,
		FS:          fsSnapshot,
	}
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
	cfg               *shared.Config
	logger            *shared.Logger
	instStore         store.InstanceStore
	resultStore       store.TaskResultStore
	deployStore       store.DeploymentStore
	svcStore          store.ServiceStore
	proxyHandler      proxy.HandleProxy
	fs                *FileSystem
	executor          *Executor
	agentTokenBackoff time.Duration

	dedupeMu sync.Mutex
	dedupe   map[string]time.Time
}

// obtainAgentTokenWithBackoff repeatedly calls fetchAgentToken using the
// bootstrap token. It performs a few quick retries and then backs off
// exponentially between attempts, using a single Syncer-level backoff value
// that is reset on successful acquisition.
func (s *Syncer) obtainAgentTokenWithBackoff(ctx context.Context, bootstrapToken string) (string, error) {
	const (
		maxBackoff   = 12 * time.Hour
		startBackoff = time.Minute
	)

	for attempt := 0; ; attempt++ {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		if attempt > 0 {
			s.logger.Debug("syncer: retrying agent token fetch", "attempt", attempt)
		}

		token, err := s.fetchAgentToken(ctx, bootstrapToken)
		if err == nil && strings.TrimSpace(token) != "" {
			s.agentTokenBackoff = 0
			s.logger.Debug("syncer: agent token obtained", "attempt", attempt)
			return token, nil
		}

		s.logger.Debug("syncer: agent token fetch failed", "attempt", attempt, "error", err)

		if attempt < 3 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(2 * time.Second):
			}
			continue
		}

		if s.agentTokenBackoff <= 0 {
			s.agentTokenBackoff = startBackoff
		} else {
			s.agentTokenBackoff *= 2
			if s.agentTokenBackoff > maxBackoff {
				s.agentTokenBackoff = maxBackoff
			}
		}

		sleep := jitter(s.agentTokenBackoff)
		s.logger.Debug("syncer: applying exponential backoff", "backoff_ms", sleep.Milliseconds(), "attempt", attempt)
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(sleep):
		}
	}
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

func NewSyncer(cfg *shared.Config, logger *shared.Logger, instStore store.InstanceStore, resStore store.TaskResultStore, deployStore store.DeploymentStore, svcStore store.ServiceStore, proxyHandler proxy.HandleProxy, handler http.Handler, auth pkgAuth.Authenticator, fs *FileSystem) *Syncer {
	return &Syncer{
		cfg:          cfg,
		logger:       logger,
		instStore:    instStore,
		resultStore:  resStore,
		deployStore:  deployStore,
		svcStore:     svcStore,
		proxyHandler: proxyHandler,
		fs:           fs,
		executor:     NewExecutor(logger, handler, instStore, auth),
		dedupe:       make(map[string]time.Time),
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
		s.logger.Debug("syncer: instance not registered; skipping WS connect")
		return nil
	}

	// Enrich context and attach instance_id explicitly for logging
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
	clientCert, err := s.ensureClientCertificate(inst.InstanceID)
	if err != nil {
		return fmt.Errorf("failed to ensure client certificate: %w", err)
	}

	if len(clientCert.Certificate) > 0 {
		if parsed, err := x509.ParseCertificate(clientCert.Certificate[0]); err == nil {
			logger.Info("syncer: client certificate ready", "cn", parsed.Subject.CommonName, "not_after", parsed.NotAfter.Format(time.RFC3339))
		}
	}

	logger.Info("syncer: publishing client certificate")
	if err := s.publishClientCertificate(ctx, inst.InstanceID, accessToken, clientCert); err != nil {
		var httpErr *httpError
		if errors.As(err, &httpErr) && (httpErr.StatusCode == http.StatusUnauthorized || httpErr.StatusCode == http.StatusForbidden) {
			logger.Warn("syncer: access token rejected; refreshing and retrying", "status", httpErr.StatusCode)
			if err := s.instStore.SetAccessToken(ctx, ""); err != nil {
				logger.Error("syncer: failed to clear invalid access token", "error", err)
			} else if accessToken, err = s.ensureAccessToken(ctx, bootstrapToken); err == nil {
				if err := s.publishClientCertificate(ctx, inst.InstanceID, accessToken, clientCert); err == nil {
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
	tlsConfig, err := s.buildPinnedTLSConfig(clientCert)
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

	defer conn.Close(websocket.StatusNormalClosure, "")

	// Set message size limit
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

	// send hello; non-blocking handshake
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
		_ = s.sendWSMessage(connCtx, conn, wsMessage{ID: ulid.Make().String(), TS: time.Now(), Kind: "hello", Hello: h})
	}

	pending, err := s.resultStore.ListUnsent(ctx)
	if err == nil && len(pending) > 0 {
		ids := make([]string, 0, len(pending))
		for _, r := range pending {
			ids = append(ids, r.ID)
		}
		logger.Debug("syncer: sending pending acks", "count", len(ids))
		if err := s.sendWSMessage(ctx, conn, wsMessage{ID: ulid.Make().String(), TS: time.Now(), Kind: "ack", IDs: ids}); err == nil {
			s.resultStore.MarkSynced(ctx, ids)
		}
	}

	logger.Debug("syncer: sending initial pull")
	if err := s.sendWSMessage(connCtx, conn, wsMessage{ID: ulid.Make().String(), TS: time.Now(), Kind: "pull"}); err != nil {
		return fmt.Errorf("failed to send initial pull: %w", err)
	}

	interval := shared.SanitizeSyncInterval(s.cfg.SyncInterval)
	go func() {
		for {
			select {
			case <-connCtx.Done():
				return
			case <-time.After(jitter(interval)):
				logger.Debug("syncer: sending periodic pull")
				if err := s.sendWSMessage(connCtx, conn, wsMessage{ID: ulid.Make().String(), TS: time.Now(), Kind: "pull"}); err != nil {
					logger.Error("syncer: failed to send periodic pull", "error", err)
					return
				}
			}
		}
	}()

	updateInterval := shared.SanitizeSyncInterval(s.cfg.SyncInterval)
	go func() {
		for {
			select {
			case <-connCtx.Done():
				return
			case <-time.After(jitter(updateInterval)):
				upd := buildAgentUpdate(connCtx, s.cfg, inst.InstanceID, s.instStore, s.deployStore, s.svcStore, s.proxyHandler, s.fs)
				msg := wsMessage{
					ID:     ulid.Make().String(),
					TS:     time.Now(),
					Kind:   "update",
					Update: upd,
				}
				if err := s.sendWSMessage(connCtx, conn, msg); err != nil {
					logger.Error("syncer: failed to send update", "error", err)
					return
				}
			}
		}
	}()

	for {
		var msg wsMessage
		if err := wsjson.Read(connCtx, conn, &msg); err != nil {
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

func (s *Syncer) buildPinnedTLSConfig(clientCert tls.Certificate) (*tls.Config, error) {
	var pool *x509.CertPool // nil => use system roots

	if s.cfg.WSCertPath != "" {
		b, err := os.ReadFile(s.cfg.WSCertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read pinned cert from %s: %w", s.cfg.WSCertPath, err)
		}
		p := x509.NewCertPool()
		if !p.AppendCertsFromPEM(b) {
			return nil, fmt.Errorf("failed to parse pinned cert from %s", s.cfg.WSCertPath)
		}
		pool = p
	} else if baseWSCACertPEM != "" {
		p := x509.NewCertPool()
		if !p.AppendCertsFromPEM([]byte(baseWSCACertPEM)) {
			return nil, fmt.Errorf("failed to parse embedded WebSocket CA cert")
		}
		pool = p
	}

	return &tls.Config{
		RootCAs:      pool, // nil => system roots
		Certificates: []tls.Certificate{clientCert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func (s *Syncer) ensureClientCertificate(instanceID string) (tls.Certificate, error) {
	certPath, keyPath := defaultClientCertPaths()

	if fileExists(certPath) && fileExists(keyPath) {
		return tls.LoadX509KeyPair(certPath, keyPath)
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("generate ecdsa key: %w", err)
	}

	serial, err := crand.Int(crand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		serial = big.NewInt(time.Now().UnixNano())
	}

	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: fmt.Sprintf("dployr-instance:%s", strings.TrimSpace(instanceID))},
		NotBefore:             time.Now().Add(-5 * time.Minute),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(crand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("create certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("marshal private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})

	dir := filepath.Dir(certPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return tls.Certificate{}, fmt.Errorf("mkdir %s: %w", dir, err)
	}

	if err := os.WriteFile(certPath, certPEM, 0o644); err != nil {
		return tls.Certificate{}, fmt.Errorf("write cert: %w", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		return tls.Certificate{}, fmt.Errorf("write key: %w", err)
	}

	return tls.X509KeyPair(certPEM, keyPEM)
}

func defaultClientCertPaths() (certPath, keyPath string) {
	var dir string
	switch runtime.GOOS {
	case "windows":
		dir = filepath.Join(os.Getenv("PROGRAMDATA"), "dployr")
	case "darwin":
		dir = "/usr/local/etc/dployr"
	default:
		dir = "/var/lib/dployrd"
	}
	return filepath.Join(dir, "client.crt"), filepath.Join(dir, "client.key")
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

// publishClientCertificate registers or rotates the client certificate with base
// using the agent access token. It first tries POST and falls back to PUT on
// conflict.
func (s *Syncer) publishClientCertificate(ctx context.Context, instanceID, agentToken string, cert tls.Certificate) error {
	if len(cert.Certificate) == 0 {
		return fmt.Errorf("client certificate is empty")
	}

	parsed, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("failed to parse client certificate: %w", err)
	}

	// Compute SPKI SHA-256 fingerprint in base64
	hash := sha256.Sum256(parsed.RawSubjectPublicKeyInfo)
	spki := base64.StdEncoding.EncodeToString(hash[:])

	body := map[string]any{
		"pem":         string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: parsed.Raw})),
		"spki_sha256": spki,
		"subject":     parsed.Subject.String(),
		"not_after":   parsed.NotAfter.Format(time.RFC3339Nano),
	}

	base := strings.TrimRight(s.cfg.BaseURL, "/")
	if base == "" {
		return fmt.Errorf("base_url is not configured")
	}

	url := fmt.Sprintf("%s/v1/agent/cert?instanceName=%s", base, instanceID)

	if err := s.sendCertRequest(ctx, http.MethodPost, url, agentToken, body); err != nil {
		var httpErr *httpError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusConflict {
			// Cert already exists; attempt rotation via PUT request
			return s.sendCertRequest(ctx, http.MethodPut, url, agentToken, body)
		}
		return err
	}

	return nil
}

type httpError struct {
	StatusCode int
	Msg        string
}

func (e *httpError) Error() string { return e.Msg }

func (s *Syncer) sendCertRequest(ctx context.Context, method, url, agentToken string, body map[string]any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal cert payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to build cert request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(agentToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cert request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return &httpError{StatusCode: resp.StatusCode, Msg: fmt.Sprintf("cert request returned status %d", resp.StatusCode)}
}

func (s *Syncer) sendWSMessage(ctx context.Context, conn *websocket.Conn, msg wsMessage) error {
	return wsjson.Write(ctx, conn, msg)
}

func (s *Syncer) handleTasks(ctx context.Context, conn *websocket.Conn, items []syncTask, logger *shared.Logger) {
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
			Payload: t.Payload,
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

		// NEW: Send task_response immediately after execution
		if err := s.sendTaskResponse(ctx, conn, t.ID, result); err != nil {
			tlog.Error("failed to send task_response", "error", err, "task_id", t.ID)
		} else {
			tlog.Debug("sent task_response", "task_id", t.ID, "success", result.Status == "done")
		}

		completedResults = append(completedResults, result)
		completedIDs = append(completedIDs, t.ID)

		// Only mark non-log-stream tasks as seen for deduplication
		if !isLogStream {
			s.markTaskSeen(t.ID)
		}
	}

	if err := s.resultStore.SaveResults(ctx, fromTaskResults(completedResults)); err != nil {
		logger.Error("failed to persist completed task results", "error", err)
		return
	}

	if len(completedIDs) > 0 {
		if err := s.sendWSMessage(ctx, conn, wsMessage{ID: ulid.Make().String(), TS: time.Now(), Kind: "ack", IDs: completedIDs}); err != nil {
			logger.Error("failed to send ack", "error", err)
		}
	}
}

// sendTaskResponse sends a task_response message back to the server for realtime requests
func (s *Syncer) sendTaskResponse(ctx context.Context, conn *websocket.Conn, taskID string, result *tasks.Result) error {
	success := result.Status == "done"

	msg := wsMessage{
		ID:        ulid.Make().String(),
		TS:        time.Now(),
		Kind:      "task_response",
		TaskID:    taskID,
		RequestID: taskID, // requestId equals taskId per strategy doc
		Success:   success,
		Data:      result.Metadata,
	}

	if success && result.Result != nil {
		msg.Data = result.Result
	}

	if !success && result.Error != "" {
		msg.TaskError = &TaskError{
			Code:    mapErrorToCode(result.Error),
			Message: result.Error,
		}
	}

	return s.sendWSMessage(ctx, conn, msg)
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

// fetchAgentToken exchanges bootstrap token for a short-lived agent
// access token that can be used to authenticate daemon operations.
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
