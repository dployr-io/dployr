package server

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"dployr.io/pkg/logger"
	"dployr.io/pkg/models"
	"dployr.io/pkg/repository"
	"dployr.io/pkg/scripts"
)

type ServerSetup struct {
	Connection  *ConnectionPool
	Project     *models.Project
	Logger      *logger.Logger
	SetupState  *SetupState
	SetupPhases []SetupPhase
}

type SetupPhase struct {
	Name    string
	Execute func(s *ServerSetup) error
}


func (s *ServerSetup) loadSetupState(ctx context.Context, r *repository.EventRepo, t *SetupState, cfg *HostConfig, conn *ConnectionPool) error {
	if _, err := os.Stat("/opt/dployr/setup.state"); err == nil {
		// Load previous setup configuration
		data, err := os.ReadFile("/opt/dployr/setup.state")
		if err != nil {
			return errors.New("failed to read setup state file")
		}

		// Parse and validate the setup state
		if err := json.Unmarshal(data, &t); err != nil {
			return errors.New("invalid setup state file format")
		}

		// Identify incomplete or failed setup steps
		if err := s.checkIncompleteSteps(ctx, r, t, cfg, conn); err != nil {
			for _, phase := range s.SetupPhases {
				phase.Execute(s)
			}
		}

		return nil
	} else if os.IsNotExist(err) {
		// Initialize fresh setup state tracking
		initialState := NewSetupState(s.Project)

		// Create setup state file
		data, err := json.Marshal(initialState)
		if err != nil {
			return errors.New("failed to marshal initial setup state")
		}

		// Ensure directory exists
		if err := os.MkdirAll("/opt/dployr", 0755); err != nil {
			return errors.New("failed to create setup directory")
		}

		// Write initial state file
		if err := os.WriteFile("/opt/dployr/setup.state", data, 0644); err != nil {
			return errors.New("failed to write initial setup state")
		}

		return nil
	} else if os.IsPermission(err) {
		return errors.New("permission denied accessing setup state file")
	} else {
		return errors.New("unknown error loading setup state")
	}
}

func (s *ServerSetup) checkIncompleteSteps(ctx context.Context, r *repository.EventRepo, t *SetupState, cfg *HostConfig, conn *ConnectionPool) error {
	if t.Status == models.Success {
		return nil
	}

	err := errors.New("incomplete setup steps detected")

	s.Logger.Warn(ctx, r, models.Setup, err.Error())
	// Check if all required phases are completed

	if t.Phases.FireWall.Status != models.Success {
		err = errors.New("required setup phase not completed: Firewall")
		s.SetupPhases = append(s.SetupPhases, SetupPhase{
			Name: "Firewall",
			Execute: func(s *ServerSetup) error {
				return s.setupFirewall(ctx, r, t, cfg, conn)
			},
		})
		s.Logger.Warn(ctx, r, models.Setup, err.Error())
		return err
	}

	if t.Phases.SystemDeps.Status != models.Success {
		err = errors.New("required setup phase not completed: System Dependencies")
		s.SetupPhases = append(s.SetupPhases, SetupPhase{
			Name: "System Dependencies",
			Execute: func(s *ServerSetup) error {
				return s.setupSystemDeps(ctx, r, t, cfg, conn)
			},
		})
		s.Logger.Warn(ctx, r, models.Setup, err.Error())
		return err
	}

	if t.Phases.Services.Status != models.Success {
		err = errors.New("required setup phase not completed: Services")
		s.SetupPhases = append(s.SetupPhases, SetupPhase{
			Name: "Services",
			Execute: func(s *ServerSetup) error {
				return s.setupServices(ctx, r, t, cfg, conn)
			},
		})
	}

	return nil
}

func (s *ServerSetup) executeSetupPhase(ctx context.Context, r *repository.EventRepo, cfg *HostConfig, conn *ConnectionPool, phaseName string, commands []string) error {
	s.Logger.Info(ctx, r, models.Setup, "Setting up "+phaseName+"...")

	err := conn.BatchExecuteCommand(ctx, r, cfg, commands, s.Logger)
	if err != nil {
		return err
	}

	return nil
}

func (s *ServerSetup) setupFirewall(ctx context.Context, r *repository.EventRepo, t *SetupState, cfg *HostConfig, conn *ConnectionPool) error {
	t.Phases.FireWall.Status = models.Pending
	err := s.executeSetupPhase(ctx, r, cfg, conn, "firewall", scripts.FirewallSetupScript)
	if err != nil {
		t.Phases.FireWall.Status = models.Failed
		return err
	}
	t.Phases.FireWall.Status = models.Success
	return nil
}

func (s *ServerSetup) setupSystemDeps(ctx context.Context, r *repository.EventRepo, t *SetupState, cfg *HostConfig, conn *ConnectionPool) error {
	t.Phases.SystemDeps.Status = models.Pending
	err := s.executeSetupPhase(ctx, r, cfg, conn, "system dependencies", scripts.SystemDepsSetupScript)
	if err != nil {
		t.Phases.SystemDeps.Status = models.Failed
		return err
	}
	t.Phases.SystemDeps.Status = models.Success
	return nil
}

func (s *ServerSetup) setupServices(ctx context.Context, r *repository.EventRepo, t *SetupState, cfg *HostConfig, conn *ConnectionPool) error {
	t.Phases.Services.Status = models.Pending
	err := s.executeSetupPhase(ctx, r, cfg, conn, "services", scripts.ServicesSetupScript)
	if err != nil {
		t.Phases.Services.Status = models.Failed
		return err
	}
	t.Phases.Services.Status = models.Success
	return nil
}
