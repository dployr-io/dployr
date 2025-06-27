package main

import (
	"context"
	"log"
	"os"

	"dployr.io/pkg/api/auth"
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

	// Create queue manager first
	qm, err := queue.NewQueueManager()
	if err != nil {
		log.Fatalf("Failed to create queue manager: %v", err)
	}

	// Start queue manager in background
	go func() {
		if err := qm.Start(context.Background()); err != nil {
			log.Printf("Queue error: %v", err)
		}
	}()

	// Create auth with queue manager
	auth := auth.InitAuth(projectRepo, eventRepo, qm)

	// Create router and run
	r := router.New(auth, qm)
	r.Run(":" + port)
}
