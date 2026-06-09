package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/BurntSushi/toml"
)

type Auth struct {
	SessionCookie string    `toml:"session_cookie,omitempty"`
	AccessToken   string    `toml:"access_token,omitempty"`
	RefreshToken  string    `toml:"refresh_token,omitempty"`
	ExpiresAt     time.Time `toml:"expires_at,omitempty"`
	UserEmail     string    `toml:"user_email,omitempty"`
}

type Config struct {
	BaseURL         string `toml:"base_url"`
	ActiveCluster   string `toml:"active_cluster,omitempty"`
	ActiveClusterID string `toml:"active_cluster_id,omitempty"`
	Auth            Auth   `toml:"auth"`
}

// configPathOverride is used in tests to redirect reads/writes to a temp file.
var configPathOverride string

func Path() string {
	if configPathOverride != "" {
		return configPathOverride
	}
	switch runtime.GOOS {
	case "darwin":
		return "/usr/local/etc/dployr/config.toml"
	case "windows":
		return filepath.Join(os.Getenv("PROGRAMDATA"), "dployr", "config.toml")
	default:
		return "/etc/dployr/config.toml"
	}
}

func Load() (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(Path(), &cfg); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("could not read config: %w", err)
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	path := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("could not open config: %w", err)
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(c)
}

func (c *Config) IsAuthenticated() bool {
	return c.Auth.SessionCookie != "" || c.Auth.AccessToken != ""
}

func (c *Config) ClearAuth() {
	c.Auth = Auth{}
}
