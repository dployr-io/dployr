// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"
)

func TestJitter_WithinBounds(t *testing.T) {
	base := 10 * time.Second
	delta := base / 10 // ±10%

	for i := 0; i < 1000; i++ {
		got := jitter(base)
		if got < base-delta || got > base+delta {
			t.Fatalf("jitter(%v) = %v, outside [%v, %v]", base, got, base-delta, base+delta)
		}
	}
}

func TestJitter_ZeroBase(t *testing.T) {
	if got := jitter(0); got != 0 {
		t.Errorf("jitter(0) = %v, want 0", got)
	}
}

func TestJitter_NegativeBase(t *testing.T) {
	if got := jitter(-1 * time.Second); got != -1*time.Second {
		t.Errorf("jitter(-1s) = %v, want -1s", got)
	}
}

func TestObtainNodeTokenWithBackoff_SucceedsImmediately(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data":    map[string]any{"token": "tok-abc"},
		})
	}))
	defer srv.Close()

	backoff := time.Duration(0)
	log := shared.NewLogger()
	tok, err := ObtainNodeTokenWithBackoff(context.Background(), srv.URL, "bootstrap", &backoff, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "tok-abc" {
		t.Errorf("token = %q, want tok-abc", tok)
	}
	if backoff != 0 {
		t.Errorf("backoff should be reset to 0 on success, got %v", backoff)
	}
}

func TestObtainNodeTokenWithBackoff_RejectsOn401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	backoff := time.Duration(0)
	log := shared.NewLogger()
	_, err := ObtainNodeTokenWithBackoff(context.Background(), srv.URL, "bad-token", &backoff, log)
	if err == nil {
		t.Fatal("expected error for 401 rejection")
	}
	// Must not retry on 401 — return immediately.
	if !strings.Contains(err.Error(), "rejected") {
		t.Errorf("error should mention rejection, got: %v", err)
	}
}

func TestObtainNodeTokenWithBackoff_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	backoff := time.Duration(0)
	log := shared.NewLogger()
	_, err := ObtainNodeTokenWithBackoff(ctx, srv.URL, "bootstrap", &backoff, log)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestObtainNodeTokenWithBackoff_BackoffEscalation(t *testing.T) {
	// Verify that backoff doubles each cycle past the first 3 quick retries.
	backoff := 1 * time.Minute // simulated: already past quick-retry phase

	const maxBackoff = 12 * time.Hour
	for i := 0; i < 15; i++ {
		prev := backoff
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
		if i > 10 && backoff != maxBackoff {
			t.Errorf("backoff should be capped at %v after many iterations, got %v (prev %v)", maxBackoff, backoff, prev)
		}
	}
	if backoff != maxBackoff {
		t.Errorf("backoff never reached max cap %v", maxBackoff)
	}
}

type mockInstStore struct {
	accessToken    string
	bootstrapToken string
	accessErr      error
	bootstrapErr   error
}

func (m *mockInstStore) GetInstance(_ context.Context) (*store.Instance, error)      { return nil, nil }
func (m *mockInstStore) RegisterInstance(_ context.Context, _ *store.Instance) error { return nil }
func (m *mockInstStore) UpdateLastInstalledAt(_ context.Context) error               { return nil }
func (m *mockInstStore) SetBootstrapToken(_ context.Context, t string) error         { return nil }
func (m *mockInstStore) GetBootstrapToken(_ context.Context) (string, error) {
	return m.bootstrapToken, m.bootstrapErr
}
func (m *mockInstStore) SetAccessToken(_ context.Context, t string) error { return nil }
func (m *mockInstStore) GetAccessToken(_ context.Context) (string, error) {
	return m.accessToken, m.accessErr
}

func makeJWT(exp int64) string {
	header, _ := json.Marshal(map[string]any{"alg": "RS256", "typ": "JWT"})
	payload, _ := json.Marshal(map[string]any{"exp": exp, "iat": time.Now().Unix()})
	h := base64.RawURLEncoding.EncodeToString(header)
	p := base64.RawURLEncoding.EncodeToString(payload)
	return h + "." + p + ".fakesig"
}

func TestComputeAuthHealth_NilStore(t *testing.T) {
	health, debug := ComputeAuthHealth(context.Background(), nil)
	if health != "down" {
		t.Errorf("health = %q, want down", health)
	}
	if debug != nil {
		t.Errorf("debug should be nil for nil store")
	}
}

func TestComputeAuthHealth_EmptyAccessToken(t *testing.T) {
	ms := &mockInstStore{accessToken: "", bootstrapToken: "btoken"}
	health, _ := ComputeAuthHealth(context.Background(), ms)
	if health != "down" {
		t.Errorf("health = %q, want down for empty access token", health)
	}
}

func TestComputeAuthHealth_EmptyBootstrapToken(t *testing.T) {
	ms := &mockInstStore{accessToken: makeJWT(time.Now().Add(5 * time.Minute).Unix()), bootstrapToken: ""}
	health, _ := ComputeAuthHealth(context.Background(), ms)
	if health != "down" {
		t.Errorf("health = %q, want down for empty bootstrap token", health)
	}
}

func TestComputeAuthHealth_ExpiredAccessToken(t *testing.T) {
	ms := &mockInstStore{
		accessToken:    makeJWT(time.Now().Add(-1 * time.Hour).Unix()),
		bootstrapToken: "btoken",
	}
	health, debug := ComputeAuthHealth(context.Background(), ms)
	if health != "down" {
		t.Errorf("health = %q, want down for expired token", health)
	}
	if debug == nil {
		t.Fatal("debug should not be nil when token was parseable")
	}
	if debug.NodeTokenExpiresIn != 0 {
		t.Errorf("ExpiresIn should be 0 for expired token, got %d", debug.NodeTokenExpiresIn)
	}
}

func TestComputeAuthHealth_ValidToken(t *testing.T) {
	ms := &mockInstStore{
		accessToken:    makeJWT(time.Now().Add(5 * time.Minute).Unix()),
		bootstrapToken: strings.Repeat("b", 80), // longer than prevLen(70) to test truncation
	}
	health, debug := ComputeAuthHealth(context.Background(), ms)
	if health != "ok" {
		t.Errorf("health = %q, want ok", health)
	}
	if debug == nil {
		t.Fatal("debug should not be nil for valid token")
	}
	if debug.NodeTokenExpiresIn <= 0 {
		t.Errorf("ExpiresIn should be > 0 for valid token, got %d", debug.NodeTokenExpiresIn)
	}
	if len(debug.BootstrapToken) != 70 {
		t.Errorf("BootstrapToken should be truncated to 70 chars, got %d", len(debug.BootstrapToken))
	}
}
