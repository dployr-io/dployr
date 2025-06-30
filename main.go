package main

import (
	"context"
	"log"
	"os"

	"dployr.io/pkg/api/auth"
	"dployr.io/pkg/api/platform"
	"dployr.io/pkg/api/router"
	"dployr.io/pkg/config"
	"dployr.io/pkg/logger"
	"dployr.io/pkg/queue"
	"dployr.io/pkg/server"

	_ "modernc.org/sqlite"
)

func main() {
	projectRepo, eventRepo := config.InitDB()

	server.NewConnectionPool()

	port := os.Getenv("PORT")
	if port == "" {
		port = "7879"
	}

	sseManager := platform.NewSSEManager()

	logger := logger.NewLogger("app", sseManager)

	// Create queue manager
	qm, err := queue.NewQueueManager(eventRepo, logger)
	if err != nil {
		log.Fatalf("Failed to create queue manager: %v", err)
	}

	// Start queue manager in background
	go func() {
		if err := qm.Start(context.Background()); err != nil {
			log.Printf("Error starting queue manager: %v", err)
		}
	}()

	auth := auth.InitAuth(projectRepo, eventRepo, qm)

	// Create router and run
	r := router.New(&router.Router{
		Auth:       auth,
		QM:         qm,
		SSEManager: sseManager,
	})
	r.Run(":" + port)
}
