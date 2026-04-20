// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/dployr-io/dployr/pkg/core/proxy"
	"github.com/dployr-io/dployr/pkg/core/service"
	"github.com/dployr-io/dployr/pkg/core/utils"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"

	"github.com/dployr-io/dployr/internal/deploy"
	"github.com/dployr-io/dployr/internal/svc_runtime"
)

type Worker struct {
	maxConcurrent int
	logger        *shared.Logger
	depsStore     store.DeploymentStore
	svcStore      store.ServiceStore
	proxyAPI      proxy.HandleProxy
	cfg           *shared.Config
	semaphore     *shared.Semaphore
	activeJobs    map[string]bool
	jobsMux       sync.RWMutex
	queue         chan string
}

// New creates a new Worker instance
func New(m int, c *shared.Config, l *shared.Logger, d store.DeploymentStore, s store.ServiceStore, p proxy.HandleProxy) *Worker {
	return &Worker{
		maxConcurrent: m,
		logger:        l,
		depsStore:     d,
		svcStore:      s,
		proxyAPI:      p,
		cfg:           c,
		semaphore:     shared.NewSemaphore(m),
		activeJobs:    make(map[string]bool),
		queue:         make(chan string, 100),
	}
}

func (w *Worker) Start(ctx context.Context) {
	for {
		select {
		case id := <-w.queue:
			if w.isRunning(id) {
				w.logger.Info("deployment with " + id + " already running, skipping")
				continue
			}

			if err := w.semaphore.Acquire(ctx); err != nil {
				w.logger.Warn("failed to acquire semaphore slot", "error", err)
				continue
			}
			w.markActive(id)

			go w.execute(ctx, id)

		case <-ctx.Done():
			w.logger.Info("Worker shutting down")
			return
		}
	}
}

func (w *Worker) Submit(id string) {
	w.queue <- id
}

func (w *Worker) execute(ctx context.Context, id string) {
	defer func() {
		w.markInactive(id)
		w.semaphore.Release()
	}()

	w.depsStore.UpdateDeploymentStatus(ctx, id, string(store.StatusInProgress))
	if err := w.runDeployment(ctx, id); err != nil {
		w.logger.Error("deployment failed", "error", err)
		w.depsStore.UpdateDeploymentStatus(ctx, id, string(store.StatusFailed))
		return
	}

	w.depsStore.UpdateDeploymentStatus(ctx, id, string(store.StatusCompleted))
}

func (w *Worker) runDeployment(ctx context.Context, id string) error {
	dataDir := utils.GetDataDir()
	logPath := filepath.Join(dataDir, ".dployr", "logs") + "/"

	d, err := w.depsStore.GetDeployment(ctx, id)
	if err != nil {
		msg := fmt.Sprintf("could not resolve user home directory: %v", err)
		w.logger.Error(msg)
		return err
	}

	shared.LogInfoF(id, logPath, "creating workspace")
	workingDir, err := deploy.SetupDir(d.Blueprint.Name)
	dir := ""
	if err != nil {
		err = fmt.Errorf("failed to setup working directory %s: %v", workingDir, err)
		shared.LogErrF(id, logPath, err)
		return err
	}

	if d.Blueprint.Source == "remote" {
		shared.LogInfoF(id, logPath, "cloning repository")
		err = deploy.CloneRepo(d.Blueprint.Remote, workingDir, d.Blueprint.WorkingDir, w.cfg)
		if err != nil {
			err = fmt.Errorf("failed to clone repository: %s", err)
			shared.LogErrF(id, logPath, err)
			return err
		}
	} else {
		shared.LogInfoF(id, logPath, "pulling image")
		err = deploy.PullImage(d.Blueprint.Image, d.Blueprint.WorkingDir, w.cfg)
		if err != nil {
			err = fmt.Errorf("failed to pull image: %s", err)
			shared.LogErrF(id, logPath, err)
			return err
		}
	}

	// Set the working directory for the service
	dir = fmt.Sprint(workingDir, "/", d.Blueprint.WorkingDir)

	svcName := utils.FormatName(d.Blueprint.Name)

	shared.LogInfoF(id, logPath, "checking for existing service")
	s, err := svc_runtime.SvcRuntime()
	if err != nil {
		err = fmt.Errorf("failed to default a compatible runtime manager: %s", err)
		shared.LogErrF(id, logPath, err)
		return err
	}

	status, err := s.Status(svcName)
	if err == nil {
		// Service exists, remove it first
		shared.LogWarnF(id, logPath, fmt.Sprintf("previous version of %s exists", svcName))
		shared.LogInfoF(id, logPath, "uninstalling previous version...")
		if status == string(service.SvcRunning) {
			s.Stop(svcName)
		}
		time.Sleep(100 * time.Millisecond)
		s.Remove(svcName)
	} else {
		// Service doesn't exist, new deployment
		shared.LogInfoF(id, logPath, "no existing service found, proceeding with installation")
	}

	bp := store.Blueprint{
		Name:       svcName,
		Desc:       d.Blueprint.Desc,
		Source:     d.Blueprint.Source,
		Type:       d.Blueprint.Type,
		Runtime:    d.Blueprint.Runtime,
		Remote:     d.Blueprint.Remote,
		RunCmd:     d.Blueprint.RunCmd,
		BuildCmd:   d.Blueprint.BuildCmd,
		Port:       d.Blueprint.Port,
		WorkingDir: dir,
		StaticDir:  d.Blueprint.StaticDir,
		Image:      d.Blueprint.Image,
		EnvVars:    d.Blueprint.EnvVars,
		Secrets:    d.Blueprint.Secrets,
		Status:     d.Blueprint.Status,
		ProjectID:  d.Blueprint.ProjectID,
	}

	shared.LogInfoF(id, logPath, "deploying application")
	err = deploy.DeployApp(bp, id, logPath)
	if err != nil {
		err = fmt.Errorf("deployment failed: %s", err)
		shared.LogErrF(id, logPath, err)
		return err
	}

	req := &store.Service{
		ID:             svcName,
		Name:           d.Blueprint.Name,
		Description:    d.Blueprint.Desc,
		Source:         d.Blueprint.Source,
		Type:           d.Blueprint.Type,
		Runtime:        d.Blueprint.Runtime.Type,
		RuntimeVersion: d.Blueprint.Runtime.Version,
		RunCmd:         d.Blueprint.RunCmd,
		BuildCmd:       d.Blueprint.BuildCmd,
		Port:           d.Blueprint.Port,
		WorkingDir:     d.Blueprint.WorkingDir,
		StaticDir:      d.Blueprint.StaticDir,
		Image:          d.Blueprint.Image,
		EnvVars:        d.Blueprint.EnvVars,
		Secrets:        d.Blueprint.Secrets,
		Remote:         d.Blueprint.Remote.Url,
		Branch:         d.Blueprint.Remote.Branch,
		CommitHash:     d.Blueprint.Remote.CommitHash,
		DeploymentId:   d.ID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	w.logger.Info("saving service", "source", req.Source, "type", req.Type)

	_, err = w.svcStore.SaveService(ctx, req)
	if err != nil {
		err = fmt.Errorf("failed to save service: %s", err)
		shared.LogErrF(id, logPath, err)
		return err
	}

	shared.LogInfoF(id, logPath, fmt.Sprintf("successfully deployed %s", d.Blueprint.Name))

	if err := w.registerProxyRoute(ctx, req); err != nil {
		w.logger.Warn("failed to register proxy route for service", "service", req.Name, "error", err)
	}

	w.depsStore.UpdateDeploymentStatus(ctx, id, string(store.StatusCompleted))

	return nil
}

func (w *Worker) registerProxyRoute(ctx context.Context, svc *store.Service) error {
	if w.proxyAPI == nil {
		return nil
	}

	serviceName := utils.FormatName(svc.Name)
	serviceDomain := serviceName + ".dployr.dev"

	upstream := fmt.Sprintf("localhost:%d", svc.Port)
	if svc.Port == 0 {
		return fmt.Errorf("port is not set")
	}

	app := proxy.App{
		Domain:   serviceDomain,
		Upstream: upstream,
		Template: proxy.TemplateReverseProxy,
	}

	w.logger.Info("registering proxy route", "domain", serviceDomain, "upstream", upstream)

	apps := map[string]proxy.App{serviceDomain: app}
	if err := w.proxyAPI.Add(apps); err != nil {
		return fmt.Errorf("failed to add proxy route: %w", err)
	}

	return nil
}

func (w *Worker) isRunning(id string) bool {
	w.jobsMux.RLock()
	defer w.jobsMux.RUnlock()
	return w.activeJobs[id]
}

func (w *Worker) markActive(id string) {
	w.jobsMux.Lock()
	defer w.jobsMux.Unlock()
	w.activeJobs[id] = true
}

func (w *Worker) markInactive(id string) {
	w.jobsMux.Lock()
	defer w.jobsMux.Unlock()
	delete(w.activeJobs, id)
}

// ActiveJobs returns the current count of active jobs
func (w *Worker) ActiveJobs() int {
	w.jobsMux.RLock()
	defer w.jobsMux.RUnlock()
	return len(w.activeJobs)
}
