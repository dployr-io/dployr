package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/dployr-io/dployr/internal/cli/config"
)

// EnvToken is the environment variable checked before config for CI usage.
const EnvToken = "DPLOYR_TOKEN"

// Client is a typed HTTP client for the dployr-base API.
type Client struct {
	http    *http.Client
	cfg     *config.Config
	cluster string // active cluster override (from flag or config)
}

func New(cfg *config.Config) *Client {
	return &Client{
		http:    &http.Client{},
		cfg:     cfg,
		cluster: cfg.ActiveCluster,
	}
}

// WithCluster returns a copy of the client scoped to a specific cluster name.
func (c *Client) WithCluster(name string) *Client {
	cp := *c
	cp.cluster = name
	return &cp
}

// Cluster returns the active cluster name for this client.
func (c *Client) Cluster() string { return c.cluster }

// do performs an HTTP request and returns the raw response.
// Callers are responsible for closing resp.Body on success.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	rawURL := strings.TrimRight(c.cfg.BaseURL, "/") + "/v1" + path
	if len(query) > 0 {
		rawURL += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	c.setAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", rawURL, err)
	}

	return resp, nil
}

func (c *Client) setAuth(req *http.Request) {
	// DPLOYR_TOKEN env var takes precedence — clean CI path.
	if token := os.Getenv(EnvToken); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		return
	}
	if c.cfg.Auth.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.Auth.AccessToken)
		return
	}
	if c.cfg.Auth.SessionCookie != "" {
		req.Header.Set("Cookie", c.cfg.Auth.SessionCookie)
	}
}

// get performs a GET and decodes JSON into T.
func get[T any](ctx context.Context, c *Client, path string, query url.Values) (T, error) {
	resp, err := c.do(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		var zero T
		return zero, err
	}
	return decodeResponse[T](resp)
}

// post performs a POST and decodes JSON into T.
func post[T any](ctx context.Context, c *Client, path string, body any) (T, error) {
	resp, err := c.do(ctx, http.MethodPost, path, nil, body)
	if err != nil {
		var zero T
		return zero, err
	}
	return decodeResponse[T](resp)
}

// patch performs a PATCH and decodes JSON into T.
func patch[T any](ctx context.Context, c *Client, path string, body any) (T, error) {
	resp, err := c.do(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		var zero T
		return zero, err
	}
	return decodeResponse[T](resp)
}

// del performs a DELETE, expecting 204 No Content.
func del(ctx context.Context, c *Client, path string) error {
	resp, err := c.do(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		return nil
	}
	return readAPIError(resp)
}

// postNoContent performs a POST expecting 204 No Content.
func postNoContent(ctx context.Context, c *Client, path string, body any) error {
	resp, err := c.do(ctx, http.MethodPost, path, nil, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		return nil
	}
	return readAPIError(resp)
}

// apiEnvelope mirrors the server's { "success": true, "data": T } wrapper.
type apiEnvelope[T any] struct {
	Data T `json:"data"`
}

// paginatedItems mirrors the server's paginated response data field.
type paginatedItems[T any] struct {
	Items []T `json:"items"`
}

func decodeResponse[T any](resp *http.Response) (T, error) {
	defer resp.Body.Close()
	var zero T
	if resp.StatusCode >= 400 {
		return zero, readAPIError(resp)
	}
	var envelope apiEnvelope[T]
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return zero, fmt.Errorf("decode response (status %d): %w", resp.StatusCode, err)
	}
	return envelope.Data, nil
}

// apiErrorEnvelope mirrors { "success": false, "error": { "message": "...", "code": "..." } }.
type apiErrorEnvelope struct {
	Error struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
}

func readAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("not authenticated — run 'dployr auth login' first")
	}
	var envelope apiErrorEnvelope
	if json.Unmarshal(body, &envelope) == nil && envelope.Error.Message != "" {
		return fmt.Errorf("%s (HTTP %d)", envelope.Error.Message, resp.StatusCode)
	}
	return fmt.Errorf("server returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
}

// RawResponse returns the raw *http.Response for callers that need to handle
// streaming or custom decoding (e.g. WebSocket upgrade).
func (c *Client) RawResponse(ctx context.Context, method, path string, query url.Values, body any) (*http.Response, error) {
	return c.do(ctx, method, path, query, body)
}
