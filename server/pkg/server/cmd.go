package server

import (
	"bytes"
	"context"
	"time"

	"golang.org/x/crypto/ssh"

	"dployr.io/pkg/logger"
	"dployr.io/pkg/models"
	"dployr.io/pkg/repository"
)

func (p *ConnectionPool) ExecuteCommand(cfg *HostConfig, command string) (*CommandResult, error) {
	client, err := p.Get(cfg)
	if err != nil {
		return nil, err
	}

	// Create session from client
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	start := time.Now()
	err = session.Run(command)
	duration := time.Since(start)

	res := &CommandResult{
		Command:  command,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	if err != nil {
		if exitError, ok := err.(*ssh.ExitError); ok {
			res.ExitCode = exitError.ExitStatus()
		} else {
			res.ExitCode = -1
		}
		res.Error = err
	}

	return res, nil
}

func (p *ConnectionPool) BatchExecuteCommand(ctx context.Context, r *repository.EventRepo, cfg *HostConfig, cmds []string, logger *logger.Logger) error {
	for _, cmd := range cmds {
		res, _ := p.ExecuteCommand(cfg, cmd)

		if res.Error != nil {
			logger.Error(ctx, r, models.Setup, res.Stderr)
			return res.Error
		}

		logger.Info(ctx, r, models.Setup, res.Stdout)
	}

	return nil
}
