package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/pressly/goose"
	"golang.org/x/oauth2"

	"dployr.io/pkg/repository"
)

const (
	Version     = "0.1.0"
	APIEndpoint = "api.dployr.io"
)

type Config struct {
	Version     string                    `yaml:"version"`
	APIEndpoint string                    `yaml:"api_endpoint"`
	Projects    map[string]ProjectConfig  `yaml:"projects"`
	Providers   map[string]ProviderConfig `yaml:"providers"`
	SSHKeys     map[string]string         `yaml:"ssh_keys"`
	Settings    UserSettings              `yaml:"settings"`
}

type ProjectConfig struct {
	ID          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	GitRepo     string            `yaml:"git_repo"`
	Domain      string            `yaml:"domain"`
	Provider    string            `yaml:"provider"`
	Environment map[string]string `yaml:"environment,omitempty"`
	CreatedAt   time.Time         `yaml:"created_at"`
}

type ProviderConfig struct {
	Type        string            `yaml:"type"` // "digitalocean", "ssh"
	Region      string            `yaml:"region,omitempty"`
	Size        string            `yaml:"size,omitempty"`
	Credentials map[string]string `yaml:"credentials"` // encrypted references
	Metadata    map[string]string `yaml:"metadata,omitempty"`
}

type UserSettings struct {
	DefaultProvider string `yaml:"default_provider,omitempty"`
	LogLevel        string `yaml:"log_level"`
	AutoSSL         bool   `yaml:"auto_ssl"`
	Notifications   bool   `yaml:"notifications"`
}

type NextJSProject struct {
	Path         string            `json:"path"`
	PackageJSON  PackageJSON       `json:"package_json"`
	NextConfig   NextConfig        `json:"next_config"`
	BuildCommand string            `json:"build_command"`
	OutputMode   string            `json:"output_mode"` // "standalone", "static"
	HasAppDir    bool              `json:"has_app_dir"`
	Dependencies map[string]string `json:"dependencies"`
}

type PackageJSON struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Scripts      map[string]string `json:"scripts"`
	Dependencies map[string]string `json:"dependencies"`
	DevDeps      map[string]string `json:"devDependencies"`
}

type NextConfig struct {
	Output       string                 `json:"output"`
	Experimental map[string]interface{} `json:"experimental"`
	Images       map[string]interface{} `json:"images"`
}

type BuildResult struct {
	Success     bool          `json:"success"`
	Command     string        `json:"command"`
	Stdout      string        `json:"stdout"`
	Stderr      string        `json:"stderr"`
	Duration    time.Duration `json:"duration"`
	PackagePath string        `json:"package_path,omitempty"`
	BuildSize   int64         `json:"build_size_bytes"`
	Error       error         `json:"error,omitempty"`
}

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: Could not load .env file:", err)
	}
}

func GetSupabaseProjectID() string { return os.Getenv("SUPABASE_PROJECT_ID") }

func GetSupabaseAnonKey() string { return os.Getenv("SUPABASE_ANON_KEY") }

func runMigrations(db *sqlx.DB) {
	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatalf("goose: %v", err)
	}

	if err := goose.Up(db.DB, "./db/migrations"); err != nil {
		log.Printf("Warning: migrations encountered issues: %v", err)
	}
}

func InitDB() (*repository.Project, *repository.Event) {
	dsn := "file:./data.sqlite3?_pragma=journal_mode(WAL)"
	db, err := sqlx.Open("sqlite", dsn)

	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	runMigrations(db)

	fmt.Println("Database initialized successfully")

	projectRepo := repository.NewProjectRepo(db)
	eventRepo := repository.NewEventRepo(db)

	return projectRepo, eventRepo
}

func GetOauth2Provider() *oidc.Provider {
	provider, err := oidc.NewProvider(
		context.Background(),
		"https://"+os.Getenv("AUTH0_DOMAIN")+"/",
	)
	if err != nil {
		log.Fatal("Failed to initialize OAuth2 provider:", err)
	}

	return provider
}

func GetOauth2Config(host string) *oauth2.Config {
	scheme := "http"
	if host != "localhost" && !strings.Contains(host, "localhost") {
		scheme = "https"
	}

	redirectURL := fmt.Sprintf("%s://%s/v1/callback", scheme, host)

	return &oauth2.Config{
		ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
		ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
		RedirectURL:  redirectURL,
		Endpoint:     GetOauth2Provider().Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}
}
