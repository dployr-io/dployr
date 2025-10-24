package main

import (
	"context"
	"dployr/pkg/core"
	"dployr/pkg/shared"
	"log"
	"os"
	"os/signal"
	"syscall"

	_auth "dployr/internal/auth"
	"dployr/internal/db"
	_store "dployr/internal/store"
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

	logger := shared.NewLogger()
	// ds := store.NewDeploymentStore(db)
	us := _store.NewUserStore(conn)
	ds := _store.NewDeploymentStore(conn)
	ctx := context.Background()
	ctx = shared.WithRequest(ctx, ulid.Make().String())
	ctx = shared.WithTrace(ctx, ulid.Make().String())

	authService := _auth.NewAuth(cfg)
	ah := auth.NewAuthHandler(us, logger, authService)
	am := auth.NewMiddleware(authService, us)

	deployer := core.NewDeployer(cfg, logger, ds)
	dh := core.NewDeploymentHandler(deployer, logger)

	// web server for API
	wh := web.WebHandler{
		AuthM: am,
		AuthH: ah,
		DepsH: dh,
	}
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
