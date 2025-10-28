package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"dployr/internal/deploy"
	"dployr/internal/svc_runtime"
	"dployr/pkg/core/service"
	"dployr/pkg/core/utils"
	"dployr/pkg/shared"
	"dployr/pkg/store"
)

type Worker struct {
	maxConcurrent int
	logger        *slog.Logger
	depsStore         store.DeploymentStore
	svcStore store.ServiceStore
	cfg        *shared.Config
	semaphore     chan struct{}
	activeJobs    map[string]bool
	jobsMux       sync.RWMutex
	queue         chan string
}

// New creates a new Worker instance
func New(m int, c *shared.Config, l *slog.Logger, d store.DeploymentStore, s store.ServiceStore) *Worker {
	return &Worker{
		maxConcurrent: m,
		logger:        l,
		depsStore:         d,
		svcStore: s,
		cfg:        c,
		semaphore:     make(chan struct{}, m),
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

			w.semaphore <- struct{}{}
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
		<-w.semaphore
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not resolve user home directory: %v", err)
	}
	logPath := homeDir + "/.dployr/logs/" + id

	d, err := w.depsStore.GetDeployment(ctx, id)
	if err != nil {
		msg := fmt.Sprintf("could not resolve user home directory: %v", err)
		w.logger.Error(msg)
		return err
	}

	shared.LogInfoF(id, logPath, "creating workspace")
	workingDir, err := deploy.SetupDir(d.Cfg.Name)
	dir := ""
	if err != nil {
		err = fmt.Errorf("failed to setup working directory %s: %v", workingDir, err)
		shared.LogErrF(id, logPath, err)
		return err
	}

	shared.LogInfoF(id, logPath, "cloning repository")
	err = deploy.CloneRepo(d.Cfg.Remote, workingDir, d.Cfg.WorkingDir, w.cfg)
	if err != nil {
		err = fmt.Errorf("failed to clone repository: %s", err)
		shared.LogErrF(id, logPath, err)
		return err
	}

	// Set the working directory for the service
	dir = fmt.Sprint(workingDir, "/", d.Cfg.WorkingDir)

	shared.LogInfoF(id, logPath, "installing runtime")
	err = deploy.SetupRuntime(d.Cfg.Runtime, dir)
	if err != nil {
		err = fmt.Errorf("failed to setup runtime: %s", err)
		shared.LogErrF(id, logPath, err)
		return err
	}

	if d.Cfg.BuildCmd != "" {
		shared.LogInfoF(id, logPath, "installing dependencies")
		err := deploy.InstallDeps(d.Cfg.BuildCmd, dir, d.Cfg.Runtime)
		if err != nil {
			err = fmt.Errorf("failed to setup working directory: %s", err)
			shared.LogErrF(id, logPath, err)
			return err
		}
	}

	shared.LogInfoF(id, logPath, "creating service")
	s := &svc_runtime.NSSMManager{}

	svcName := utils.FormatName(d.Cfg.Name)

	status, err := s.Status(svcName)
	if err == nil {
		shared.LogWarnF(id, logPath, fmt.Sprintf("previous version of %s exists", svcName))
		shared.LogInfoF(id, logPath, "uninstalling previous version...")
		if status == string(service.SvcRunning) {
			s.Stop(svcName)
		}
		time.Sleep(2 * time.Second)
		s.Remove(svcName)
		return nil
	}

	shared.LogInfoF(id, logPath, "creating run file")
	exe, cmdArgs, err := utils.GetExeArgs(d.Cfg.Runtime, d.Cfg.RunCmd)

	if err != nil {
		err = fmt.Errorf("%s", err)
		shared.LogErrF(id, logPath, err)
		return err
	}

	bat, err := svc_runtime.CreateRunFile(d.Cfg, dir, exe, cmdArgs)
	if err != nil {
		err = fmt.Errorf("%s", err)
		shared.LogErrF(id, logPath, err)
		return err
	}

	shared.LogInfoF(id, logPath, "installing service")
	err = s.Install(svcName, d.Cfg.Desc, bat, dir, d.Cfg.EnvVars)
	if err != nil {
		err = fmt.Errorf("service installation failed: %s", err)
		shared.LogErrF(id, logPath, err)
		return err
	}

	shared.LogInfoF(id, logPath, "starting service")
	err = s.Start(svcName)
	if err != nil {
		err = fmt.Errorf("service start failed: %s", err)
		shared.LogErrF(id, logPath, err)
		return err
	}

	req := &store.Service{
		ID: svcName,
		Name: d.Cfg.Name,
		Description: d.Cfg.Desc,
		Source: d.Cfg.Source,
		Runtime: d.Cfg.Runtime.Type,
		RuntimeVersion: d.Cfg.Runtime.Version,
		RunCmd: d.Cfg.RunCmd,
		BuildCmd: d.Cfg.BuildCmd,
		Port: d.Cfg.Port,
		WorkingDir: d.Cfg.WorkingDir,
		StaticDir: d.Cfg.StaticDir,
		Image: d.Cfg.Image,
		EnvVars: d.Cfg.EnvVars,
		Remote: d.Cfg.Remote,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	w.svcStore.CreateService(ctx, req)

	shared.LogInfoF(id, logPath, fmt.Sprintf("successfully deployed %s", d.Cfg.Name))
	w.depsStore.UpdateDeploymentStatus(ctx, id, string(store.StatusCompleted))

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
