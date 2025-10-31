package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	_auth "dployr/internal/auth"
	"dployr/internal/db"
	_deploy "dployr/internal/deploy"
	_proxy "dployr/internal/proxy"
	_service "dployr/internal/service"
	_store "dployr/internal/store"
	_stream "dployr/internal/stream"
	"dployr/internal/web"
	"dployr/internal/worker"
	"dployr/pkg/auth"
	"dployr/pkg/core/deploy"
	"dployr/pkg/core/proxy"
	"dployr/pkg/core/service"
	"dployr/pkg/core/stream"
	"dployr/pkg/shared"

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
	ss := _store.NewServiceStore(conn, ds)
	ctx := context.Background()
	ctx = shared.WithRequest(ctx, ulid.Make().String())
	ctx = shared.WithTrace(ctx, ulid.Make().String())

	w := worker.New(5, cfg, logger, ds, ss) // 5 concurrent deployments

	authService := _auth.Init(cfg)
	ah := auth.NewAuthHandler(us, logger, authService)
	am := auth.NewMiddleware(authService, us)

	api := _deploy.Init(cfg, logger, ds, w)
	deployer := deploy.NewDeployer(cfg, logger, ds, api)
	dh := deploy.NewDeploymentHandler(deployer, logger)

	_services := _service.Init(cfg, logger, ss)
	servicer := service.NewServicer(cfg, logger, ss, _services)
	sh := service.NewServiceHandler(servicer, logger)

	proxyState := _proxy.LoadState()
	ps := _proxy.Init(proxyState)
	proxier := proxy.NewProxier(proxyState, ps)
	ph := proxy.NewProxyHandler(proxier, logger)

	logsService := _stream.Init()
	homeDir, _ := os.UserHomeDir()
	logsDir := filepath.Join(homeDir, ".dployr", "logs")
	ls := stream.NewLogStreamer(logsDir, logsService)
	lh := stream.NewLogStreamHandler(ls, logger)

	wh := web.WebHandler{
		AuthH:  ah,
		DepsH:  dh,
		SvcH:   sh,
		LogsH:  lh,
		ProxyH: ph,
		AuthM:  am,
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
