package main

import (
	"context"
	"dployr/pkg/core/deploy"
	"dployr/pkg/shared"
	"log"
	"os"
	"os/signal"
	"syscall"

	_auth "dployr/internal/auth"
	"dployr/internal/db"
	_store "dployr/internal/store"
	_deploy "dployr/internal/deploy"
	"dployr/internal/web"
	"dployr/internal/worker"
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
	us := _store.NewUserStore(conn)
	ds := _store.NewDeploymentStore(conn)
	ctx := context.Background()
	ctx = shared.WithRequest(ctx, ulid.Make().String())
	ctx = shared.WithTrace(ctx, ulid.Make().String())

	w := worker.New(5, cfg, logger, ds) // 5 concurrent deployments

	authService := _auth.NewAuth(cfg)
	ah := auth.NewAuthHandler(us, logger, authService)
	am := auth.NewMiddleware(authService, us)

	api := _deploy.New(cfg, logger, ds, w)

	deployer := deploy.NewDeployer(cfg, logger, ds, api)
	dh := deploy.NewDeploymentHandler(deployer, logger)

	wh := web.WebHandler{
		AuthH: ah,
		DepsH: dh,
		AuthM: am,
	}

	go func() {
		if err := wh.NewServer(cfg.Port); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	go func() {
		w.Start(ctx)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("shutting down gracefully...")
}
