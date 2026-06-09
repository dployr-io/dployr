package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dployr-io/dployr/internal/cli/config"
)

func TestExchangeOIDCToken_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/auth/oidc/exchange" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["token"] == "" {
			t.Error("expected token in request body")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"sessionId": "test-session-id"},
		})
	}))
	defer srv.Close()

	c := New(&config.Config{BaseURL: srv.URL})
	sessionId, err := c.ExchangeOIDCToken(context.Background(), "eyJhbGciOiJSUzI1NiJ9.payload.sig")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessionId != "test-session-id" {
		t.Errorf("sessionId = %q, want test-session-id", sessionId)
	}
}

func TestExchangeOIDCToken_401Rejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := New(&config.Config{BaseURL: srv.URL})
	_, err := c.ExchangeOIDCToken(context.Background(), "invalid-token")
	if err == nil {
		t.Error("expected error for 401 response")
	}
}

func TestExchangeOIDCToken_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New(&config.Config{BaseURL: srv.URL})
	_, err := c.ExchangeOIDCToken(context.Background(), "some-token")
	if err == nil {
		t.Error("expected error for 500 response")
	}
}
