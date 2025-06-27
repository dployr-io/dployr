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

type SetupUserAccountArgs struct {
	Ctx         context.Context
	EventRepo   *repository.Event
	ProjectRepo *repository.Project
	models.SetupUserData
	SetupState *SetupState
	Cfg        *HostConfig
	Conn       *ConnectionPool
}

func (s *ServerSetup) SetupUserAccount(args SetupUserAccountArgs) (*models.Project, error) {
	project, err := args.ProjectRepo.GetByID(args.Ctx, args.UserID)
	if err != nil {
		return nil, err
	}
	if project != nil {
		return project, nil
	}

	s.Logger.Info(args.Ctx, args.EventRepo, models.Setup, "Setting up user account for "+args.Name)

	project, err = args.ProjectRepo.Create(args.Ctx, &models.Project{
		ID:   args.UserID,
		Name: args.Name + "'s Project",
	})

	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ServerSetup) LoadSetupState(args SetupUserAccountArgs) error {
	if _, err := os.Stat("/opt/dployr/setup.state"); err == nil {
		// Load previous setup configuration
		data, err := os.ReadFile("/opt/dployr/setup.state")
		if err != nil {
			return errors.New("failed to read setup state file")
		}

		// Parse and validate the setup state
		if err := json.Unmarshal(data, &args.SetupState); err != nil {
			return errors.New("invalid setup state file format")
		}

		// Identify incomplete or failed setup steps
		if err := s.checkIncompleteSteps(args); err != nil {
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

func (s *ServerSetup) checkIncompleteSteps(args SetupUserAccountArgs) error {
	if args.SetupState.Status == models.Success {
		return nil
	}

	err := errors.New("incomplete setup steps detected")

	s.Logger.Warn(args.Ctx, args.EventRepo, models.Setup, err.Error())

	if args.SetupState.Phases.FireWall.Status != models.Success {
		err = errors.New("required setup phase not completed: Firewall")
		s.SetupPhases = append(s.SetupPhases, SetupPhase{
			Name: "Firewall",
			Execute: func(s *ServerSetup) error {
				return s.setupFirewall(args)
			},
		})
		s.Logger.Warn(args.Ctx, args.EventRepo, models.Setup, err.Error())
		return err
	}

	if args.SetupState.Phases.SystemDeps.Status != models.Success {
		err = errors.New("required setup phase not completed: System Dependencies")
		s.SetupPhases = append(s.SetupPhases, SetupPhase{
			Name: "System Dependencies",
			Execute: func(s *ServerSetup) error {
				return s.setupSystemDeps(args)
			},
		})
		s.Logger.Warn(args.Ctx, args.EventRepo, models.Setup, err.Error())
		return err
	}

	if args.SetupState.Phases.Services.Status != models.Success {
		err = errors.New("required setup phase not completed: Services")
		s.SetupPhases = append(s.SetupPhases, SetupPhase{
			Name: "Services",
			Execute: func(s *ServerSetup) error {
				return s.setupServices(args)
			},
		})
		s.Logger.Warn(args.Ctx, args.EventRepo, models.Setup, err.Error())
		return err
	}

	return nil
}

func (s *ServerSetup) executeSetupPhase(args SetupUserAccountArgs, phaseName string, commands []string) error {
	s.Logger.Info(args.Ctx, args.EventRepo, models.Setup, "Setting up "+phaseName+"...")

	err := args.Conn.BatchExecuteCommand(args.Ctx, args.EventRepo, args.Cfg, commands, s.Logger)
	if err != nil {
		return err
	}

	return nil
}

func (s *ServerSetup) setupFirewall(args SetupUserAccountArgs) error {
	args.SetupState.Phases.FireWall.Status = models.Pending
	err := s.executeSetupPhase(args, "firewall", scripts.FirewallSetupScript)
	if err != nil {
		args.SetupState.Phases.FireWall.Status = models.Failed
		return err
	}
	args.SetupState.Phases.FireWall.Status = models.Success
	return nil
}

func (s *ServerSetup) setupSystemDeps(args SetupUserAccountArgs) error {
	args.SetupState.Phases.SystemDeps.Status = models.Pending
	err := s.executeSetupPhase(args, "system dependencies", scripts.SystemDepsSetupScript)
	if err != nil {
		args.SetupState.Phases.SystemDeps.Status = models.Failed
		return err
	}
	args.SetupState.Phases.SystemDeps.Status = models.Success
	return nil
}

func (s *ServerSetup) setupServices(args SetupUserAccountArgs) error {
	args.SetupState.Phases.Services.Status = models.Pending
	err := s.executeSetupPhase(args, "services", scripts.ServicesSetupScript)
	if err != nil {
		args.SetupState.Phases.Services.Status = models.Failed
		return err
	}
	args.SetupState.Phases.Services.Status = models.Success
	return nil
}
