// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Address      string
	Port         int
	MaxWorkers   int
	BaseURL      string
	SyncInterval time.Duration
	WSCertPath   string
	TaskDedupTTL time.Duration
}

func LoadConfig() (*Config, error) {
	configPath := getSystemConfigPath()
	err := LoadTomlFile(configPath)

	if err != nil {
		return nil, err
	}

	return &Config{
		Address:      getEnv("ADDRESS", "localhost"),
		Port:         getEnvAsInt("PORT", 7879),
		MaxWorkers:   getEnvAsInt("MAX_WORKERS", 5),
		BaseURL:      getEnv("BASE_URL", ""),
		SyncInterval: getEnvAsDuration("SYNC_INTERVAL", 30*time.Second),
		WSCertPath:   getEnv("WS_CERT_PATH", ""),
		TaskDedupTTL: getEnvAsPositiveDuration("TASK_DEDUP_TTL", 5*time.Minute),
	}, nil
}

func LoadTomlFile(path string) error {
	var config map[string]any
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return err
	}

	for key, value := range config {
		// Convert kebab-case/snake_case to UPPER_CASE for system vars
		envKey := strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
		os.Setenv(envKey, fmt.Sprintf("%v", value))
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

const (
	minSyncInterval = 5 * time.Second
	maxSyncInterval = 30 * time.Second
)

// getEnvAsDuration returns a sanitized duration from an environment variable.
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return sanitizeSyncInterval(defaultValue)
	}
	if d, err := time.ParseDuration(value); err == nil {
		return sanitizeSyncInterval(d)
	}
	if intValue, err := strconv.Atoi(value); err == nil {
		return sanitizeSyncInterval(time.Duration(intValue) * time.Second)
	}
	return sanitizeSyncInterval(defaultValue)
}

// getEnvAsPositiveDuration returns a positive duration from an environment variable.
func getEnvAsPositiveDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	if d, err := time.ParseDuration(value); err == nil && d > 0 {
		return d
	}
	if intValue, err := strconv.Atoi(value); err == nil && intValue > 0 {
		return time.Duration(intValue) * time.Second
	}
	return defaultValue
}

// SanitizeSyncInterval clamps a duration to safe sync bounds.
func SanitizeSyncInterval(v time.Duration) time.Duration {
	if v <= 0 {
		return 30 * time.Second
	}
	if v < minSyncInterval {
		return minSyncInterval
	}
	if v > maxSyncInterval {
		return maxSyncInterval
	}
	return v
}

func sanitizeSyncInterval(v time.Duration) time.Duration {
	return SanitizeSyncInterval(v)
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func GetToken() (string, error) {
	configDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not resolve user home directory: %v", err)
	}
	configPath := configDir + "/.dployr/token.json"

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("could not read config file: %v. Please run 'dployr login' first", err)
	}

	var config map[string]string
	if err := json.Unmarshal(configData, &config); err != nil {
		return "", fmt.Errorf("could not parse config file: %v", err)
	}

	accessToken, exists := config["access_token"]
	if !exists {
		return "", fmt.Errorf("no access token found in config. Please run 'dployr login' first")
	}

	if isTokenExpired(accessToken) {
		refreshToken, refreshExists := config["refresh_token"]
		if !refreshExists {
			return "", fmt.Errorf("access token expired and no refresh token available. Please run 'dployr login' again")
		}

		newAccessToken, err := refreshAccessToken(refreshToken)
		if err != nil {
			return "", fmt.Errorf("failed to refresh access token: %v. Please run 'dployr login' again", err)
		}

		config["access_token"] = newAccessToken
		updatedConfigData, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal updated config: %v", err)
		}

		if err := os.WriteFile(configPath, updatedConfigData, 0600); err != nil {
			return "", fmt.Errorf("failed to save updated token: %v", err)
		}

		return newAccessToken, nil
	}

	return accessToken, nil
}

// isTokenExpired checks if a JWT token is expired or will expire soon (within 1 minute)
func isTokenExpired(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return true
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return true
	}

	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return true
	}

	return time.Now().Unix() >= (claims.Exp - 60)
}

// refreshAccessToken uses the refresh token to get a new access token
func refreshAccessToken(refreshToken string) (string, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %v", err)
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		return "", fmt.Errorf("BASE_URL is not configured; CLI must talk to base, not the local daemon")
	}

	reqBody := map[string]string{
		"refresh_token": refreshToken,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal refresh request: %v", err)
	}

	refreshURL := strings.TrimRight(baseURL, "/") + "/auth/refresh"
	resp, err := http.Post(refreshURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("refresh failed with status: %d", resp.StatusCode)
	}

	var refreshResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&refreshResp); err != nil {
		return "", fmt.Errorf("failed to parse refresh response: %v", err)
	}

	return refreshResp.AccessToken, nil
}

// system-wide, accessible to all users
func getSystemConfigPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("PROGRAMDATA"), "dployr", "config.toml")
	case "darwin":
		return "/usr/local/etc/dployr/config.toml"
	default:
		return "/etc/dployr/config.toml"
	}
}
