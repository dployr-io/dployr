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
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	coresystem "dployr/pkg/core/system"
	"dployr/pkg/shared"
	"dployr/pkg/store"
	"dployr/pkg/tasks"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type syncerCtxKey string

const ctxKeyInstanceID syncerCtxKey = "instance_id"

const baseWSCACertPEM = ""

type Syncer struct {
	cfg               *shared.Config
	logger            *shared.Logger
	instStore         store.InstanceStore
	resultStore       store.TaskResultStore
	executor          *Executor
	agentTokenBackoff time.Duration
}

// wsMessage represents WebSocket messages exchanged with base.
type wsMessage struct {
	Kind  string     `json:"kind"`
	Items []syncTask `json:"items,omitempty"`
	IDs   []string   `json:"ids,omitempty"`
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

func NewSyncer(cfg *shared.Config, logger *shared.Logger, inst store.InstanceStore, results store.TaskResultStore, handler http.Handler) *Syncer {
	return &Syncer{
		cfg:         cfg,
		logger:      logger,
		instStore:   inst,
		resultStore: results,
		executor:    NewExecutor(logger, handler),
	}
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
		return "", fmt.Errorf("failed to obtain agent token: %w", err)
	}
	if err := s.instStore.SetAccessToken(ctx, accessToken); err != nil {
		s.logger.Error("syncer: failed to persist access token", "error", err)
	}
	s.logger.Debug("syncer: access token persisted")
	return accessToken, nil
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

	// Enrich context with instance_id for all subsequent logs
	ctx = context.WithValue(ctx, ctxKeyInstanceID, inst.InstanceID)
	logger := s.logger.WithContext(ctx)

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
		fmt.Sprintf("/v1/agent/instances/%s/ws", inst.InstanceID)

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

	logger.Info("syncer: websocket connected to base")

	pending, err := s.resultStore.ListUnsent(ctx)
	if err == nil && len(pending) > 0 {
		ids := make([]string, 0, len(pending))
		for _, r := range pending {
			ids = append(ids, r.ID)
		}
		logger.Debug("syncer: sending pending acks", "count", len(ids))
		if err := s.sendWSMessage(ctx, conn, wsMessage{Kind: "ack", IDs: ids}); err == nil {
			s.resultStore.MarkSynced(ctx, ids)
		}
	}

	logger.Debug("syncer: sending initial pull")
	if err := s.sendWSMessage(ctx, conn, wsMessage{Kind: "pull"}); err != nil {
		return fmt.Errorf("failed to send initial pull: %w", err)
	}

	for {
		var msg wsMessage
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			logger.Error("syncer: websocket read failed; will reconnect", "error", err)
			return fmt.Errorf("websocket read failed: %w", err)
		}

		logger.Debug("syncer: received message", "kind", msg.Kind)
		switch msg.Kind {
		case "task":
			logger.Debug("syncer: received tasks", "count", len(msg.Items))
			s.handleTasks(ctx, conn, msg.Items, logger)
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
		dir = "/etc/dployr"
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

	url := fmt.Sprintf("%s/v1/agent/instances/%s/cert", base, instanceID)

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
	if mode == coresystem.ModeUpdating {
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
		task := &tasks.Task{
			ID:      t.ID,
			Type:    t.Type,
			Payload: t.Payload,
			Status:  t.Status,
		}

		result := s.executor.Execute(ctx, task)
		completedResults = append(completedResults, result)
		completedIDs = append(completedIDs, t.ID)
	}

	if err := s.resultStore.SaveResults(ctx, fromTaskResults(completedResults)); err != nil {
		logger.Error("failed to persist completed task results", "error", err)
		return
	}

	if len(completedIDs) > 0 {
		if err := s.sendWSMessage(ctx, conn, wsMessage{Kind: "ack", IDs: completedIDs}); err != nil {
			logger.Error("failed to send ack", "error", err)
		}
	}
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
