package main

import (
	"context"
	"dployr/pkg/shared"
	"log"
	"os"
	"os/signal"
	"syscall"

	_auth "dployr/internal/auth"
	"dployr/internal/db"
	"dployr/internal/store"
	"dployr/internal/web"
	"dployr/pkg/auth"

	"github.com/oklog/ulid/v2"
)

func main() {
	conn, err := db.Open()
	if err != nil {
		log.Fatalf("Failed to open database connection: %s", err)
		return
	}

	cfg, err := shared.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// logger := shared.NewLogger()
	// ds := store.NewDeploymentStore(db)
	us := store.NewUserStore(conn)
	ctx := context.Background()
	ctx = shared.WithRequest(ctx, ulid.Make().String())
	ctx = shared.WithTrace(ctx, ulid.Make().String())

	// req := &core.DeployRequest{
	//     Name:        "MyApp",
	//     Description: "Simple test deploy",
	//     Runtime: models.RuntimeObj{
	// 		Type:    models.RuntimeGo,
	// 		Version: "1.24.1",
	// 	},
	//     RunCmd:      "go run main.go",
	//     Port:        8080,
	//     StaticDir:   "public",
	//     EnvVars:     map[string]string{"ENV": "prod"},
	// }

	// d := core.NewDeployer(cfg, logger, ds)
	// d.Deploy(ctx, req)

	// Initialize core service (business logic)
	// service := core.NewService(db, logger)

	// Initialize auth dependencies
	authService := _auth.NewAuth(cfg)
	authHandler := auth.NewAuthHandler(us, authService)

	// web server for API
	wh := web.WebHandler{Handler: authHandler}
	go func() {
		if err := wh.NewServer(cfg.Port); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("shutting down gracefully...")
}
