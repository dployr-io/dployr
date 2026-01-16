package auth

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/dployr-io/dployr/pkg/core/system"
	"github.com/dployr-io/dployr/pkg/store"
)

type agentTokenResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

type HTTPError struct {
	StatusCode int
	Msg        string
}

func (e *HTTPError) Error() string { return e.Msg }

func FetchAgentToken(ctx context.Context, baseURL, bootstrapToken string) (string, error) {
	base := strings.TrimRight(baseURL, "/")
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

func ObtainAgentTokenWithBackoff(ctx context.Context, baseURL, bootstrapToken string, backoffDuration *time.Duration) (string, error) {
	const (
		maxBackoff   = 12 * time.Hour
		startBackoff = time.Minute
	)

	for attempt := 0; ; attempt++ {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		token, err := FetchAgentToken(ctx, baseURL, bootstrapToken)
		if err == nil && strings.TrimSpace(token) != "" {
			*backoffDuration = 0
			return token, nil
		}

		if attempt < 3 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(2 * time.Second):
			}
			continue
		}

		if *backoffDuration <= 0 {
			*backoffDuration = startBackoff
		} else {
			*backoffDuration *= 2
			if *backoffDuration > maxBackoff {
				*backoffDuration = maxBackoff
			}
		}

		sleep := jitter(*backoffDuration)
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(sleep):
		}
	}
}

func PublishClientCertificate(ctx context.Context, baseURL, instanceID, agentToken string, cert tls.Certificate) error {
	if len(cert.Certificate) == 0 {
		return fmt.Errorf("client certificate is empty")
	}

	parsed, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("failed to parse client certificate: %w", err)
	}

	spki, err := ComputeCertFingerprint(cert)
	if err != nil {
		return err
	}

	body := map[string]any{
		"pem":         string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: parsed.Raw})),
		"spki_sha256": spki,
		"subject":     parsed.Subject.String(),
		"not_after":   parsed.NotAfter.Format(time.RFC3339Nano),
	}

	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		return fmt.Errorf("base_url is not configured")
	}

	url := fmt.Sprintf("%s/v1/agent/cert?instanceName=%s", base, instanceID)

	if err := sendCertRequest(ctx, http.MethodPost, url, agentToken, body); err != nil {
		var httpErr *HTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusConflict {
			return sendCertRequest(ctx, http.MethodPut, url, agentToken, body)
		}
		return err
	}

	return nil
}

func sendCertRequest(ctx context.Context, method, url, agentToken string, body map[string]any) error {
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

	return &HTTPError{StatusCode: resp.StatusCode, Msg: fmt.Sprintf("cert request returned status %d", resp.StatusCode)}
}

func ComputeAuthHealth(ctx context.Context, instStore store.InstanceStore) (health string, debug *system.AuthDebug) {
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
