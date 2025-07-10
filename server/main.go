package main

import (
	"os"
	"time"

	"dployr.io/pkg/api/auth"
	"dployr.io/pkg/api/platform"
	"dployr.io/pkg/api/router"
	"dployr.io/pkg/config"
	"dployr.io/pkg/queue"
	"dployr.io/pkg/server"
)

func main() {
	projectRepo, eventRepo := config.InitDB()

	server.NewConnectionPool()

	port := os.Getenv("PORT")
	if port == "" {
		port = "7879"
	}

	// Init queue manager 
	_queue := queue.NewQueue(3, time.Second, queue.CreateHandler())

	// Init auth
	auth := auth.InitAuth(projectRepo, eventRepo, _queue)

	// Init ssh manager
	ssh := platform.NewSshManager()

	// Create router and run
	r := router.New(auth, _queue, ssh)
	r.Run(":" + port)
}
