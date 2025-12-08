// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dployr-io/dployr/pkg/auth"
	"github.com/dployr-io/dployr/pkg/core/deploy"
	"github.com/dployr-io/dployr/pkg/core/proxy"
	"github.com/dployr-io/dployr/pkg/core/service"
	"github.com/dployr-io/dployr/pkg/core/system"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/version"

	_auth "github.com/dployr-io/dployr/internal/auth"
	"github.com/dployr-io/dployr/internal/db"
	_deploy "github.com/dployr-io/dployr/internal/deploy"
	_proxy "github.com/dployr-io/dployr/internal/proxy"
	_service "github.com/dployr-io/dployr/internal/service"
	_store "github.com/dployr-io/dployr/internal/store"
	_system "github.com/dployr-io/dployr/internal/system"
	"github.com/dployr-io/dployr/internal/web"
	"github.com/dployr-io/dployr/internal/worker"
)

func main() {
	var showVersion = flag.Bool("version", false, "show version information")
	flag.Parse()

	if *showVersion {
		info := version.Get()
		fmt.Println(info.String())
		os.Exit(0)
	}

	log.Printf("Starting %s", version.Get().String())

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
	ds := _store.NewDeploymentStore(conn)
	ss := _store.NewServiceStore(conn, ds)
	is := _store.NewInstanceStore(conn)
	trs := _store.NewTaskResultStore(conn)

	ctx := context.Background()

	w := worker.New(5, cfg, logger, ds, ss) // 5 concurrent deployments

	as := _auth.Init(cfg, is)
	am := auth.NewMiddleware(as)

	api := _deploy.Init(cfg, logger, ds, w)
	deployer := deploy.NewDeployer(cfg, logger, ds, api)
	dh := deploy.NewDeploymentHandler(deployer, logger)

	services := _service.Init(cfg, logger, ss)
	servicer := service.NewServicer(cfg, logger, ss, services)
	sh := service.NewServiceHandler(servicer, logger)

	proxyState := _proxy.LoadState()
	ps := _proxy.Init(proxyState)
	proxier := proxy.NewProxier(proxyState, ps)
	ph := proxy.NewProxyHandler(proxier, logger)

	sysSvc := _system.NewDefaultService(cfg, is, trs)
	sysH := system.NewServiceHandler(sysSvc)
	metricsH := _system.NewMetrics(cfg, is, trs)

	wh := web.WebHandler{
		DepsH:    dh,
		SvcH:     sh,
		ProxyH:   ph,
		SystemH:  sysH,
		AuthM:    am,
		MetricsH: metricsH,
	}

	mux := wh.BuildMux(cfg)

	syncer := _system.NewSyncer(cfg, logger, is, trs, mux, as)

	go func() {
		if err := wh.NewServer(cfg); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	go func() {
		w.Start(ctx)
	}()

	go func() {
		syncer.Start(ctx)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("shutting down gracefully...")
}
