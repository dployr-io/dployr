package main

import (
	"os"
	"time"

	"dployr.io/pkg/api/auth"
	"dployr.io/pkg/api/middleware"
	"dployr.io/pkg/api/platform"
	"dployr.io/pkg/api/router"
	"dployr.io/pkg/config"
	"dployr.io/pkg/queue"
	"dployr.io/pkg/server"
)

func main() {
	repos := config.InitDB()

	if (repos == nil) {
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
	r := router.New(repos, _queue, ssh, rl, j)
	r.Run(":" + port)
}
