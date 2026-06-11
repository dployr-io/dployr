package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dployr-io/dployr/internal/cli/config"
)

func TestRequestEmailOTP_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/auth/login/email" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := New(&config.Config{BaseURL: srv.URL})
	if _, err := c.RequestEmailOTP(context.Background(), "user@example.com"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRequestEmailOTP_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := New(&config.Config{BaseURL: srv.URL})
	if _, err := c.RequestEmailOTP(context.Background(), "bad@example.com"); err == nil {
		t.Error("expected error for non-2xx response")
	}
}

func TestVerifyEmailOTP_NamedSessionCookie(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "sess-tok-abc", Path: "/"})
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(&config.Config{BaseURL: srv.URL})
	cookie, err := c.VerifyEmailOTP(context.Background(), "user@example.com", "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cookie == "" {
		t.Error("expected non-empty cookie string")
	}
}

func TestVerifyEmailOTP_DployrNamedCookie(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "dployr_session", Value: "dployr-sess-xyz", Path: "/"})
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(&config.Config{BaseURL: srv.URL})
	cookie, err := c.VerifyEmailOTP(context.Background(), "user@example.com", "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cookie == "" {
		t.Error("expected non-empty cookie string")
	}
}

func TestVerifyEmailOTP_SetCookieHeaderFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Set-Cookie", "other_session=raw-val; Path=/; HttpOnly")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(&config.Config{BaseURL: srv.URL})
	cookie, err := c.VerifyEmailOTP(context.Background(), "user@example.com", "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cookie == "" {
		t.Error("expected cookie from Set-Cookie fallback")
	}
}

func TestVerifyEmailOTP_NoCookieReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(&config.Config{BaseURL: srv.URL})
	_, err := c.VerifyEmailOTP(context.Background(), "user@example.com", "123456")
	if err == nil {
		t.Error("expected error when no session cookie is returned")
	}
}

func TestVerifyEmailOTP_Non200ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := New(&config.Config{BaseURL: srv.URL})
	_, err := c.VerifyEmailOTP(context.Background(), "user@example.com", "wrongcode")
	if err == nil {
		t.Error("expected error for non-200 response")
	}
}

func TestLogout_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/auth/logout" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(&config.Config{BaseURL: srv.URL})
	if err := c.Logout(context.Background()); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLogout_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New(&config.Config{BaseURL: srv.URL})
	if err := c.Logout(context.Background()); err == nil {
		t.Error("expected error for 500 response")
	}
}
