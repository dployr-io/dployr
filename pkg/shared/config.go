package shared

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Address        string
	Port           int
	SocketPath     string
	MaxWorkers     int
	ServerHome     string
	SecretKey      string
	GitHubToken    string
	GitLabToken    string
	BitBucketToken string
}

func LoadConfig() (*Config, error) {
	home, _ := os.UserHomeDir()
	_ = LoadEnvFile(filepath.Join(home, ".dployr", ".env"))

	secret := getEnv("SECRET_KEY", "")
	if secret == "" {
		return nil, fmt.Errorf("failed to load secret key")
	}

	return &Config{
		Address:        getEnv("ADDRESS", "localhost"),
		Port:           getEnvAsInt("PORT", 7879),
		SocketPath:     getEnv("SOCKET_PATH", fmt.Sprintf("%s/dployr.sock", ServerHome)),
		MaxWorkers:     getEnvAsInt("MAX_WORKERS", 5),
		ServerHome:     getEnv("SERVER_HOME", ServerHome),
		GitHubToken:    getEnv("GITHUB_TOKEN", ""),
		GitLabToken:    getEnv("GITLAB_TOKEN", ""),
		BitBucketToken: getEnv("BITBUCKET_TOKEN", ""),
		SecretKey:      secret,
	}, nil
}

func LoadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		os.Setenv(key, val)
	}

	return scanner.Err()
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
	configPath := configDir + "/.dployr/config.json"

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
