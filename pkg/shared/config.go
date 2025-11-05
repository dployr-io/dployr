package shared

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Address        string
	Port           int
	MaxWorkers     int
	Secret         string
	GitHubToken    string
	GitLabToken    string
	BitBucketToken string
}

func LoadConfig() (*Config, error) {
	configPath := getSystemConfigPath()
	err := LoadTomlFile(configPath)

	if err != nil {
		return nil, err
	}

	secret := getEnv("SECRET", "")
	if secret == "" {
		return nil, fmt.Errorf("failed to load secret")
	}

	return &Config{
		Address:        getEnv("ADDRESS", "localhost"),
		Port:           getEnvAsInt("PORT", 7879),
		MaxWorkers:     getEnvAsInt("MAX_WORKERS", 5),
		GitHubToken:    getEnv("GITHUB_TOKEN", ""),
		GitLabToken:    getEnv("GITLAB_TOKEN", ""),
		BitBucketToken: getEnv("BITBUCKET_TOKEN", ""),
		Secret:         secret,
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

	token, exists := config["token"]
	if !exists {
		return "", fmt.Errorf("no token found in config. Please run 'dployr login' first")
	}

	return token, nil
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
