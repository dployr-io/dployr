package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// systemMethodCase drives table tests for the three system POST methods.
type systemMethodCase struct {
	name     string
	call     func(*Client) error
	wantPath string
}

func systemCases(tag string) []systemMethodCase {
	return []systemMethodCase{
		{
			name:     "SystemReboot",
			call:     func(c *Client) error { return c.SystemReboot(context.Background(), tag) },
			wantPath: "/v1/instances/" + tag + "/system/reboot",
		},
		{
			name:     "SystemRestart",
			call:     func(c *Client) error { return c.SystemRestart(context.Background(), tag) },
			wantPath: "/v1/instances/" + tag + "/system/restart",
		},
	}
}

func TestSystemMethods_SendsClusterId(t *testing.T) {
	const tag = "quiet-moon"
	const clusterID = "cl-abc"

	for _, tc := range systemCases(tag) {
		t.Run(tc.name, func(t *testing.T) {
			var gotQuery string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotQuery = r.URL.RawQuery
				w.WriteHeader(http.StatusNoContent)
			}))
			defer srv.Close()

			c := newTestClient(t, srv.URL).WithCluster(clusterID)
			if err := tc.call(c); err != nil {
				t.Fatalf("%s error: %v", tc.name, err)
			}
			if !strings.Contains(gotQuery, "clusterId="+clusterID) {
				t.Errorf("%s: query = %q, want clusterId=%s", tc.name, gotQuery, clusterID)
			}
		})
	}
}

func TestSystemMethods_CorrectPath(t *testing.T) {
	const tag = "quiet-moon"

	for _, tc := range systemCases(tag) {
		t.Run(tc.name, func(t *testing.T) {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.WriteHeader(http.StatusNoContent)
			}))
			defer srv.Close()

			c := newTestClient(t, srv.URL)
			_ = tc.call(c) // ignore error — we only care about path
			if gotPath != tc.wantPath {
				t.Errorf("%s: path = %q, want %q", tc.name, gotPath, tc.wantPath)
			}
		})
	}
}

func TestSystemMethods_NoCluster_OmitsClusterId(t *testing.T) {
	const tag = "quiet-moon"

	for _, tc := range systemCases(tag) {
		t.Run(tc.name, func(t *testing.T) {
			var gotQuery string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotQuery = r.URL.RawQuery
				w.WriteHeader(http.StatusNoContent)
			}))
			defer srv.Close()

			c := newTestClient(t, srv.URL) // no cluster set
			_ = tc.call(c)
			if strings.Contains(gotQuery, "clusterId") {
				t.Errorf("%s: unexpected clusterId in query %q when no cluster is set", tc.name, gotQuery)
			}
		})
	}
}

func TestClusterQuery_ReturnsNilWhenNoCluster(t *testing.T) {
	c := newTestClient(t, "https://api.example.com")
	if q := c.clusterQuery(); q != nil {
		t.Errorf("clusterQuery() = %v, want nil when cluster is empty", q)
	}
}

func TestClusterQuery_ReturnsClusterIdWhenSet(t *testing.T) {
	c := newTestClient(t, "https://api.example.com").WithCluster("cl-xyz")
	q := c.clusterQuery()
	if q == nil {
		t.Fatal("clusterQuery() returned nil, want url.Values with clusterId")
	}
	if got := q.Get("clusterId"); got != "cl-xyz" {
		t.Errorf("clusterQuery()[clusterId] = %q, want cl-xyz", got)
	}
}
