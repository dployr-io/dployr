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
	"path/filepath"
	"syscall"

	"github.com/dployr-io/dployr/pkg/auth"
	"github.com/dployr-io/dployr/pkg/core/cluster"
	"github.com/dployr-io/dployr/pkg/core/deploy"
	"github.com/dployr-io/dployr/pkg/core/proxy"
	"github.com/dployr-io/dployr/pkg/core/service"
	"github.com/dployr-io/dployr/pkg/core/system"
	coreutils "github.com/dployr-io/dployr/pkg/core/utils"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/version"

	dockerclient "github.com/docker/docker/client"

	_auth "github.com/dployr-io/dployr/internal/auth"
	"github.com/dployr-io/dployr/internal/db"
	_deploy "github.com/dployr-io/dployr/internal/deploy"
	_proxy "github.com/dployr-io/dployr/internal/proxy"
	_service "github.com/dployr-io/dployr/internal/service"
	_storage "github.com/dployr-io/dployr/internal/storage"
	_store "github.com/dployr-io/dployr/internal/store"
	_system "github.com/dployr-io/dployr/internal/system"
	_terminal "github.com/dployr-io/dployr/internal/terminal"
	"github.com/dployr-io/dployr/internal/web"
	"github.com/dployr-io/dployr/internal/worker"
	pkgstorage "github.com/dployr-io/dployr/pkg/core/storage"
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

	proxyState := _proxy.LoadState()
	ps := _proxy.Init(proxyState, logger)

	workerMaxConcurrent := 5
	w := worker.New(workerMaxConcurrent, cfg, logger, ds, ss, is, ps)

	as := _auth.Init(cfg, is)
	am := auth.NewMiddleware(as)

	dockerCli, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		log.Printf("warning: failed to connect to Docker daemon: %v", err)
	}

	api := _deploy.Init(cfg, logger, ds, w, dockerCli)
	deployer := deploy.NewDeployer(cfg, logger, ds, api)
	dh := deploy.NewDeploymentHandler(deployer, logger)
	bh := deploy.NewBuildHandler(deployer, logger)

	proxier := proxy.NewProxier(proxyState, ps)
	ph := proxy.NewProxyHandler(proxier, logger)

	redeployFn := func(name string) error {
		ctx := context.Background()
		svc, err := ss.GetService(ctx, name)
		if err != nil || svc == nil {
			return fmt.Errorf("service %s not found: %w", name, err)
		}
		dep, err := ds.GetDeployment(ctx, svc.DeploymentId)
		if err != nil || dep == nil {
			return fmt.Errorf("no deployment found for service %s: %w", name, err)
		}
		logPath := filepath.Join(coreutils.GetDataDir(), ".dployr", "logs") + "/"
		return _deploy.DeployApp(dep.Blueprint, name, logPath, cfg, dockerCli)
	}

	services := _service.Init(cfg, logger, ss, ps, redeployFn)
	servicer := service.NewServicer(cfg, logger, ss, services)
	sh := service.NewServiceHandler(servicer, logger)

	sysSvc := _system.NewDefaultService(cfg, is, trs)
	sysH := system.NewServiceHandler(sysSvc)
	mh := _system.NewMetrics(cfg, is, trs)

	fs := _system.NewFS()
	fsH := system.NewFSHandler(fs, logger)

	topCollector := _system.NewTopCollector()
	topH := system.NewTopHandler(topCollector, logger)

	terminalH := _terminal.NewHandler(logger)

	storageMounter := _storage.NewMounter(logger)
	storageH := pkgstorage.NewHandler(storageMounter, logger)

	clusterSetup := _system.NewClusterSetup()
	clusterH := cluster.NewHandler(clusterSetup, logger)

	wh := web.WebHandler{
		DepsH:    dh,
		SvcH:     sh,
		ProxyH:   ph,
		SystemH:  sysH,
		FSH:      fsH,
		TopH:     topH,
		BuildH:   bh,
		AuthM:    am,
		MetricsH: mh,
		StorageH: storageH,
		ClusterH: clusterH,
	}

	mux := wh.BuildMux(cfg)

	// Restore cgroup slices for any clusters that survived a node reprovision.
	// On a blank-slate node this is a no-op; slices are created proactively
	// via the setup_cluster task sent at cluster assignment time.
	if cfg.ClusterMemory > 0 {
		if deployments, err := ds.ListDeployments(ctx, 1000, 0); err == nil {
			seen := make(map[string]struct{})
			for _, d := range deployments {
				id := d.Blueprint.ClusterID
				if id == "" {
					continue
				}
				if _, ok := seen[id]; ok {
					continue
				}
				seen[id] = struct{}{}
				if err := _system.EnsureClusterSlice(id, cfg.ClusterMemory, cfg.ClusterCPU); err != nil {
					logger.Warn("failed to ensure cluster slice at startup", "cluster_id", id, "error", err)
				}
			}
		}
	}

	syncer := _system.NewSyncer(cfg, logger, is, trs, ds, ss, ps, mux, as, fs, workerMaxConcurrent, w.ActiveJobs)
	w.SetCompletionHandler(func(id string) {
		syncer.RequestFullSync()
	})
	syncer.Executor().SetTerminalHandler(terminalH)

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
