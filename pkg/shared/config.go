package shared

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"dployr/pkg/core/utils"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/golang-jwt/jwt/v4"
)

type Config struct {
	Address        string
	Port           int
	MaxWorkers     int
	PrivateKey     *rsa.PrivateKey
	PublicKey      *rsa.PublicKey
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

	privateKey, publicKey, err := loadOrGenerateKeys()

	if err != nil {
		return nil, err
	}

	return &Config{
		Address:        getEnv("ADDRESS", "localhost"),
		Port:           getEnvAsInt("PORT", 7879),
		MaxWorkers:     getEnvAsInt("MAX_WORKERS", 5),
		GitHubToken:    getEnv("GITHUB_TOKEN", ""),
		GitLabToken:    getEnv("GITLAB_TOKEN", ""),
		BitBucketToken: getEnv("BITBUCKET_TOKEN", ""),
		PrivateKey:     privateKey,
		PublicKey:      publicKey,
	}, nil
}

func loadOrGenerateKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	dataDir := utils.GetDataDir()
	keyPath := filepath.Join(dataDir, "jwt_private.pem")

	// Try loading existing key
	if keyData, err := os.ReadFile(keyPath); err == nil {
		privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
		if err != nil {
			return nil, nil, err
		}
		return privateKey, &privateKey.PublicKey, nil
	}

	// Generate new 2048-bit RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Save to disk
	keyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	})

	if err := os.WriteFile(keyPath, pemData, 0600); err != nil {
		return nil, nil, err
	}

	return privateKey, &privateKey.PublicKey, nil
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

	addr := fmt.Sprintf("http://%s:%d", cfg.Address, cfg.Port)

	reqBody := map[string]string{
		"refresh_token": refreshToken,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal refresh request: %v", err)
	}

	resp, err := http.Post(addr+"/auth/refresh", "application/json", bytes.NewBuffer(jsonData))
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
