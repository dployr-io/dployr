package config

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/pressly/goose"
	sqlite3 "modernc.org/sqlite"

	"dployr.io/pkg/repository"
)

// Version is injected at build time via -ldflags (CI) or from version.txt
var Version = "dev"

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: Could not load .env file:", err)
	}

	sql.Register("sqlite3", &sqlite3.Driver{})
}

// Backward compatibility for Postgres
func GetDSN(portOverride ...string) string {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	if len(portOverride) > 0 {
		port = portOverride[0]
	}

	return "host=" + host + " user=" + user + " password=" + password + " dbname=" + dbname + " port=" + port + " sslmode=require"
}

func runMigrations(db *sqlx.DB) {
	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatalf("goose: %v", err)
	}

	if err := goose.Up(db.DB, "./db/migrations"); err != nil {
		log.Printf("Warning: migrations encountered issues: %v", err)
	}
}

func InitDB() (repos *repository.AppRepos) {
	dsn := "file:data.db?_foreign_keys=on&_journal=WAL"
	db, err := sqlx.Open("sqlite3", dsn)

	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	runMigrations(db)

	fmt.Println("Database initialized successfully")

	projectRepo := repository.NewProjectRepo(db)
	eventRepo := repository.NewEventRepo(db)
	userRepo := repository.NewUserRepo(db)
	tokenRepo := repository.NewMagicTokenRepo(db)
	deploymentRepo := repository.NewDeploymentRepo(db)
	logRepo := repository.NewLogRepo(db)
	refreshTokenRepo := repository.NewRefreshTokenRepo(db)

	return &repository.AppRepos{
		UserRepo:       userRepo,
		MagicTokenRepo: tokenRepo,
		ProjectRepo:    projectRepo,
		EventRepo:      eventRepo,
		DeploymentRepo: deploymentRepo,
		LogRepo:        logRepo,
		RefreshTokenRepo: refreshTokenRepo,
	}
}
