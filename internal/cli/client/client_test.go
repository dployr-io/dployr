package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dployr-io/dployr/internal/cli/config"
)

// newTestClient returns a Client pointing at the given test server URL.
func newTestClient(t *testing.T, apiURL string) *Client {
	t.Helper()
	cfg := &config.Config{BaseURL: apiURL}
	return New(cfg)
}

func TestSetAuth_EnvTokenTakesPrecedence(t *testing.T) {
	t.Setenv(EnvToken, "bearer-env-token")

	cfg := &config.Config{
		Auth: config.Auth{
			AccessToken:   "bearer-cfg-token",
			SessionCookie: "session=abc",
		},
	}
	c := New(cfg)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	c.setAuth(req)

	if got := req.Header.Get("Authorization"); got != "Bearer bearer-env-token" {
		t.Errorf("Authorization = %q, want env token", got)
	}
	if req.Header.Get("Cookie") != "" {
		t.Error("Cookie header should not be set when env token is present")
	}
}

func TestSetAuth_AccessTokenOverCookie(t *testing.T) {
	t.Setenv(EnvToken, "")

	cfg := &config.Config{
		Auth: config.Auth{
			AccessToken:   "bearer-cfg-token",
			SessionCookie: "session=abc",
		},
	}
	c := New(cfg)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	c.setAuth(req)

	if got := req.Header.Get("Authorization"); got != "Bearer bearer-cfg-token" {
		t.Errorf("Authorization = %q, want config access token", got)
	}
	if req.Header.Get("Cookie") != "" {
		t.Error("Cookie header should not be set when access token is present")
	}
}

func TestSetAuth_SessionCookieFallback(t *testing.T) {
	t.Setenv(EnvToken, "")

	cfg := &config.Config{
		Auth: config.Auth{SessionCookie: "session=xyz"},
	}
	c := New(cfg)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	c.setAuth(req)

	if got := req.Header.Get("Cookie"); got != "session=xyz" {
		t.Errorf("Cookie = %q, want session cookie", got)
	}
	if req.Header.Get("Authorization") != "" {
		t.Error("Authorization header should not be set when only session cookie is present")
	}
}

func TestSetAuth_NoAuth(t *testing.T) {
	t.Setenv(EnvToken, "")

	c := New(&config.Config{})
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	c.setAuth(req)

	if req.Header.Get("Authorization") != "" || req.Header.Get("Cookie") != "" {
		t.Error("No auth headers should be set when config and env are empty")
	}
}

func TestWithCluster_ReturnsNewClient(t *testing.T) {
	c := newTestClient(t, "https://api.example.com")
	c2 := c.WithCluster("prod")

	if c2 == c {
		t.Error("WithCluster must return a new pointer, not the same")
	}
	if c2.Cluster() != "prod" {
		t.Errorf("Cluster() = %q, want prod", c2.Cluster())
	}
	if c.Cluster() != "" {
		t.Errorf("Original cluster should be unchanged, got %q", c.Cluster())
	}
}

func TestWithCluster_ChainedCalls(t *testing.T) {
	c := newTestClient(t, "https://api.example.com")
	c2 := c.WithCluster("staging").WithCluster("prod")
	if c2.Cluster() != "prod" {
		t.Errorf("Cluster() = %q after chained WithCluster, want prod", c2.Cluster())
	}
}

func TestReadAPIError_401ReturnsLoginPrompt(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusUnauthorized,
		Body:       http.NoBody,
	}
	err := readAPIError(resp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not authenticated") {
		t.Errorf("error = %q, should mention 'not authenticated'", err.Error())
	}
}

func TestReadAPIError_APIMessageExtracted(t *testing.T) {
	body := `{"success":false,"error":{"message":"cluster not found","code":"resource.missing_resource"}}`
	resp := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       newReadCloser(body),
	}
	err := readAPIError(resp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "cluster not found") {
		t.Errorf("error = %q, want API message", err.Error())
	}
}

func TestReadAPIError_FallbackToRawBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       newReadCloser("internal error"),
	}
	err := readAPIError(resp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %q, want HTTP status in message", err.Error())
	}
}

// ── do — URL construction ────────────────────────────────────────────────────

func TestDo_CorrectURLAndMethod(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	resp, err := c.do(context.Background(), http.MethodGet, "/clusters", nil, nil)
	if err != nil {
		t.Fatalf("do() error: %v", err)
	}
	resp.Body.Close()

	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/v1/clusters" {
		t.Errorf("path = %q, want /v1/clusters", gotPath)
	}
}

func TestDo_QueryParamsAttached(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	q := url.Values{"cluster": {"prod"}}
	resp, err := c.do(context.Background(), http.MethodGet, "/services", q, nil)
	if err != nil {
		t.Fatalf("do() error: %v", err)
	}
	resp.Body.Close()

	if !strings.Contains(gotQuery, "cluster=prod") {
		t.Errorf("query = %q, want cluster=prod", gotQuery)
	}
}

func TestDo_JSONBodySent(t *testing.T) {
	var gotContentType string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	resp, err := c.do(context.Background(), http.MethodPost, "/auth/login/email", nil, map[string]string{"email": "a@b.com"})
	if err != nil {
		t.Fatalf("do() error: %v", err)
	}
	resp.Body.Close()

	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotContentType)
	}
	if gotBody["email"] != "a@b.com" {
		t.Errorf("body.email = %v, want a@b.com", gotBody["email"])
	}
}

func TestListClusters_UnwrapsEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Server wraps list in { "data": { "clusters": [...] } }
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"clusters": []map[string]any{
					{"id": "cl-1", "name": "prod"},
				},
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	clusters, err := c.ListClusters(context.Background())
	if err != nil {
		t.Fatalf("ListClusters error: %v", err)
	}
	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(clusters))
	}
	if clusters[0].Name != "prod" {
		t.Errorf("cluster.Name = %q, want prod", clusters[0].Name)
	}
}

func TestMe_UnwrapsEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Server wraps user in { "data": { "user": {...}, "clusters": [...] } }
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"user":     map[string]any{"id": "u-1", "email": "alice@example.com", "name": "Alice", "provider": "email"},
				"clusters": []any{},
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	me, err := c.Me(context.Background())
	if err != nil {
		t.Fatalf("Me error: %v", err)
	}
	if me.Email != "alice@example.com" {
		t.Errorf("me.Email = %q, want alice@example.com", me.Email)
	}
}

func TestListServices_UnwrapsPaginated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"items": []map[string]any{
					{"id": "svc-1", "name": "api", "clusterId": "cl-1"},
				},
				"pagination": map[string]any{"page": 1, "pageSize": 20, "totalItems": 1, "totalPages": 1},
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	services, err := c.ListServices(context.Background(), 0)
	if err != nil {
		t.Fatalf("ListServices error: %v", err)
	}
	if len(services) != 1 || services[0].Name != "api" {
		t.Errorf("ListServices = %v, want [{name:api}]", services)
	}
}

type readCloser struct{ *strings.Reader }

func (rc readCloser) Close() error { return nil }

func newReadCloser(s string) readCloser {
	return readCloser{strings.NewReader(s)}
}
