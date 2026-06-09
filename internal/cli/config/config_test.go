package config

import (
	"path/filepath"
	"testing"
	"time"
)

func TestIsAuthenticated(t *testing.T) {
	tests := []struct {
		name string
		auth Auth
		want bool
	}{
		{"empty", Auth{}, false},
		{"session cookie only", Auth{SessionCookie: "session=abc"}, true},
		{"access token only", Auth{AccessToken: "dpat_x.y"}, true},
		{"both set", Auth{SessionCookie: "session=abc", AccessToken: "dpat_x.y"}, true},
		{"email only (not enough)", Auth{UserEmail: "a@b.com"}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{Auth: tc.auth}
			if got := cfg.IsAuthenticated(); got != tc.want {
				t.Errorf("IsAuthenticated() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestClearAuth(t *testing.T) {
	cfg := &Config{
		Auth: Auth{
			SessionCookie: "session=abc",
			AccessToken:   "dpat_x.y",
			UserEmail:     "a@b.com",
			ExpiresAt:     time.Now().Add(time.Hour),
		},
	}
	cfg.ClearAuth()
	if cfg.IsAuthenticated() {
		t.Error("IsAuthenticated() should be false after ClearAuth")
	}
	if cfg.Auth != (Auth{}) {
		t.Errorf("Auth should be zero value after ClearAuth, got %+v", cfg.Auth)
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	orig := configPathOverride
	configPathOverride = path
	t.Cleanup(func() { configPathOverride = orig })

	cfg := &Config{
		BaseURL:       "https://base.dployr.io",
		ActiveCluster: "prod",
		Auth: Auth{
			AccessToken: "dpat_abc.secret",
			UserEmail:   "user@example.com",
		},
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.BaseURL != cfg.BaseURL {
		t.Errorf("BaseURL = %q, want %q", loaded.BaseURL, cfg.BaseURL)
	}
	if loaded.ActiveCluster != cfg.ActiveCluster {
		t.Errorf("ActiveCluster = %q, want %q", loaded.ActiveCluster, cfg.ActiveCluster)
	}
	if loaded.Auth.AccessToken != cfg.Auth.AccessToken {
		t.Errorf("Auth.AccessToken = %q, want %q", loaded.Auth.AccessToken, cfg.Auth.AccessToken)
	}
}

func TestLoad_MissingFile_ReturnsEmpty(t *testing.T) {
	orig := configPathOverride
	configPathOverride = "/nonexistent/path/config.toml"
	t.Cleanup(func() { configPathOverride = orig })

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() on missing file should not error, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil")
	}
}
