package main

import (
	"embed"
	"os"
	"time"

	"dployr.io/pkg/api/auth"
	"dployr.io/pkg/api/middleware"
	"dployr.io/pkg/api/platform"
	"dployr.io/pkg/api/router"
	"dployr.io/pkg/config"
	"dployr.io/pkg/queue"
	"dployr.io/pkg/server"

	_ "dployr.io/docs"
)

//go:embed public/*
var staticFiles embed.FS

//go:embed db/migrations/*
var migrationsFS embed.FS

// @title           dployr API
// @version         0.1
// @description     Your app, your server, your rules
// @termsOfService  https://dployr.io/terms

// @contact.name   API Support
// @contact.url    https://dployr.io/support
// @contact.email  support@dployr.io

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:7879
// @BasePath  /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	repos := config.InitDB(migrationsFS)

	if repos == nil {
		panic("Database failed to initialize properly")
	}

	server.NewConnectionPool()

	port := os.Getenv("PORT")
	if port == "" {
		port = "7879"
	}

	// Init queue manager
	_queue := queue.NewQueue(3, time.Second, queue.CreateHandler())

	// Init rate limiter - 15 minute window
	rl := middleware.NewRateLimiter(3, 15*time.Minute)
	middleware.CleanupRateLimit(rl)

	// Init ssh manager
	ssh := platform.NewSshManager()

	// Init JWT manager
	j := auth.NewJWTManager()

	// Create router and run
	r := router.New(repos, _queue, ssh, rl, j, staticFiles)
	r.Run(":" + port)
}
