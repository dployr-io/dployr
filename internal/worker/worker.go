// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	instStore     store.InstanceStore
	proxyAPI      proxy.HandleProxy
	cfg           *shared.Config
	semaphore     *shared.Semaphore
	activeJobs    map[string]bool
	jobsMux       sync.RWMutex
	queue         chan string
	onComplete    func(id string)
}

// New creates a new Worker instance
func New(m int, c *shared.Config, l *shared.Logger, d store.DeploymentStore, s store.ServiceStore, i store.InstanceStore, p proxy.HandleProxy) *Worker {
	return &Worker{
		maxConcurrent: m,
		logger:        l,
		depsStore:     d,
		svcStore:      s,
		instStore:     i,
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

func (w *Worker) SetCompletionHandler(fn func(id string)) {
	w.onComplete = fn
}

func (w *Worker) execute(ctx context.Context, id string) {
	defer func() {
		w.markInactive(id)
		w.semaphore.Release()
	}()

	w.depsStore.UpdateDeploymentStatus(ctx, id, string(store.StatusInProgress))
	logPath := filepath.Join(utils.GetDataDir(), ".dployr", "logs") + "/"

	name, err := w.runDeployment(ctx, id)
	if err != nil {
		w.logger.Error("deployment failed", "error", err)
		w.depsStore.UpdateDeploymentStatus(ctx, id, string(store.StatusFailed))
		w.notifyComplete(id)
		go w.submitDeploymentLogs(ctx, id, name, logPath)
		return
	}

	w.depsStore.UpdateDeploymentStatus(ctx, id, string(store.StatusCompleted))
	w.notifyComplete(id)
	go w.submitDeploymentLogs(ctx, id, name, logPath)
}

func (w *Worker) notifyComplete(id string) {
	if w.onComplete != nil {
		w.onComplete(id)
	}
}

func (w *Worker) runDeployment(ctx context.Context, id string) (string, error) {
	dataDir := utils.GetDataDir()
	logPath := filepath.Join(dataDir, ".dployr", "logs") + "/"

	d, err := w.depsStore.GetDeployment(ctx, id)
	if err != nil {
		msg := fmt.Sprintf("could not resolve user home directory: %v", err)
		w.logger.Error(msg)
		return "", err
	}

	svcName := utils.FormatName(d.Blueprint.Name)

	// Guard: source=remote deployments must only run on build nodes.
	// Check before SetupDir to avoid unnecessary filesystem operations.
	if d.Blueprint.Source == store.SourceRemote && w.cfg.Role != store.NodeRoleBuild {
		err = fmt.Errorf("instance node received source=remote deployment %s — expected source=image; possible routing error", d.ID)
		shared.LogErrF(svcName, logPath, err)
		return svcName, err
	}

	shared.LogInfoF(svcName, logPath, "creating workspace")
	workingDir, err := deploy.SetupDir(d.Blueprint.Name)
	dir := ""
	if err != nil {
		err = fmt.Errorf("failed to setup working directory %s: %v", workingDir, err)
		shared.LogErrF(svcName, logPath, err)
		return svcName, err
	}

	switch d.Blueprint.Source {
	case store.SourceImage:
		shared.LogInfoF(svcName, logPath, "pulling image")
		err = deploy.PullImage(d.Blueprint.Image, workingDir, w.cfg)
		if err != nil {
			err = fmt.Errorf("failed to pull image: %s", err)
			shared.LogErrF(svcName, logPath, err)
			return svcName, err
		}
		dir = workingDir
	case store.SourceRemote:
		shared.LogInfoF(svcName, logPath, "cloning repository")
		err = deploy.CloneRepo(d.Blueprint.Remote, workingDir, w.cfg)
		if err != nil {
			err = fmt.Errorf("failed to clone repository: %s", err)
			shared.LogErrF(svcName, logPath, err)
			return svcName, err
		}
		dir = workingDir
		if d.Blueprint.WorkingDir != "" {
			dir = filepath.Join(workingDir, d.Blueprint.WorkingDir)
		}
	default:
		err = fmt.Errorf("unknown deployment source %q", d.Blueprint.Source)
		shared.LogErrF(svcName, logPath, err)
		return svcName, err
	}

	shared.LogInfoF(svcName, logPath, "checking for existing service")
	s, err := svc_runtime.SvcRuntime()
	if err != nil {
		err = fmt.Errorf("failed to default a compatible runtime manager: %s", err)
		shared.LogErrF(svcName, logPath, err)
		return svcName, err
	}

	status, err := s.Status(svcName)
	if err == nil {
		// Service exists, remove it first
		shared.LogWarnF(svcName, logPath, fmt.Sprintf("previous version of %s exists", svcName))
		shared.LogInfoF(svcName, logPath, "uninstalling previous version...")
		if status == string(service.SvcRunning) {
			s.Stop(svcName)
		}
		time.Sleep(100 * time.Millisecond)
		s.Remove(svcName)
	} else {
		// Service doesn't exist, new deployment
		shared.LogInfoF(svcName, logPath, "no existing service found, proceeding with installation")
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

	shared.LogInfoF(svcName, logPath, "deploying application")
	err = deploy.DeployApp(bp, svcName, logPath, w.cfg)
	if err != nil {
		err = fmt.Errorf("deployment failed: %s", err)
		shared.LogErrF(svcName, logPath, err)
		return svcName, err
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

	_, err = w.svcStore.UpsertService(ctx, req)
	if err != nil {
		err = fmt.Errorf("failed to save service: %s", err)
		shared.LogErrF(svcName, logPath, err)
		return svcName, err
	}

	shared.LogInfoF(svcName, logPath, fmt.Sprintf("successfully deployed %s", d.Blueprint.Name))

	if err := w.registerProxyRoute(req); err != nil {
		w.logger.Warn("failed to register proxy route for service", "service", req.Name, "error", err)
	}

	w.depsStore.UpdateDeploymentStatus(ctx, id, string(store.StatusCompleted))

	return svcName, nil
}

func (w *Worker) registerProxyRoute(svc *store.Service) error {
	if w.proxyAPI == nil {
		return nil
	}

	serviceName := utils.FormatName(svc.Name)
	serviceDomain := serviceName + ".dployr.dev"
	url := "localhost:%d"
	upstream := ""
	port := 3000

	if svc.Port == 0 {
		upstream = fmt.Sprintf(url, port)
	} else {
		upstream = fmt.Sprintf(url, svc.Port)
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

func (w *Worker) submitDeploymentLogs(ctx context.Context, id string, name string, logPath string) {
	logs, err := w.readDeploymentLogs(name, logPath)
	if err != nil {
		w.logger.Error("failed to read deployment logs", "error", err)
		return
	}

	token, err := w.instStore.GetAccessToken(ctx)
	if err != nil {
		w.logger.Error("failed to get authentication token", "error", err)
		return
	}

	finishURL := strings.TrimRight(w.cfg.BaseURL, "/") + "/v1/deployments/finish"

	d, err := w.depsStore.GetDeployment(ctx, id)
	if err != nil {
		w.logger.Error("failed to retrieve deployment", "deployment_id", id, "error", err)
		return
	}

	if d == nil {
		w.logger.Error("deployment not found", "deployment_id", id)
		return
	}

	bp := d.Blueprint
	blueprint := map[string]any{
		"name":             bp.Name,
		"type":             string(bp.Type),
		"source":           string(bp.Source),
		"description":      bp.Desc,
		"run_cmd":          bp.RunCmd,
		"build_cmd":        bp.BuildCmd,
		"port":             bp.Port,
		"working_dir":      bp.WorkingDir,
		"static_dir":       bp.StaticDir,
		"image":            bp.Image,
		"runtime_type":     string(bp.Runtime.Type),
		"runtime_version":  string(bp.Runtime.Version),
		"remote_url":       bp.Remote.Url,
		"remote_branch":    bp.Remote.Branch,
		"remote_commit_hash": bp.Remote.CommitHash,
	}
	if d.UserId != nil {
		blueprint["user_id"] = *d.UserId
	}

	payload := map[string]any{
		"token":     token,
		"id":        id,
		"logs":      logs,
		"blueprint": blueprint,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		w.logger.Warn("failed to marshal deployment logs", "error", err)
		return
	}

	resp, err := http.Post(finishURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		w.logger.Warn("failed to submit deployment logs", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		w.logger.Warn("deployment log submission failed", "status", resp.StatusCode, "response", string(body))
		return
	}

	w.logger.Info("deployment logs submitted successfully", "deployment_id", id)
}

func (w *Worker) readDeploymentLogs(name string, logPath string) (string, error) {
	logFile := filepath.Join(logPath, strings.ToLower(name)+".log")
	data, err := os.ReadFile(logFile)
	if err != nil {
		return "", fmt.Errorf("failed to read log file: %w", err)
	}
	return string(data), nil
}
