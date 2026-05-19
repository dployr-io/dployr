// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package svc_runtime

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
)

const dockerOpTimeout = 30 * time.Second

func dockerCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), dockerOpTimeout)
}

type DockerService struct {
	cli dockerAPI
}

func (d *DockerService) Status(name string) (string, error) {
	ctx, cancel := dockerCtx()
	defer cancel()
	info, err := d.cli.ContainerInspect(ctx, name)
	if err != nil {
		return "", fmt.Errorf("service %s does not exist", name)
	}
	if info.State != nil && info.State.Running {
		return "running", nil
	}
	return "stopped", nil
}

func (d *DockerService) Install(name, desc, runCmd, workDir string, envVars map[string]string) error {
	return nil
}

func (d *DockerService) Start(name string) error {
	ctx, cancel := dockerCtx()
	defer cancel()
	if err := d.cli.ContainerStart(ctx, name, container.StartOptions{}); err != nil {
		return fmt.Errorf("docker start %s: %w", name, err)
	}
	return nil
}

func (d *DockerService) Stop(name string) error {
	ctx, cancel := dockerCtx()
	defer cancel()
	if err := d.cli.ContainerStop(ctx, name, container.StopOptions{}); err != nil {
		return fmt.Errorf("docker stop %s: %w", name, err)
	}
	return nil
}

func (d *DockerService) Remove(name string) error {
	ctx, cancel := dockerCtx()
	defer cancel()
	if err := d.cli.ContainerRemove(ctx, name, container.RemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("docker rm -f %s: %w", name, err)
	}
	return nil
}

func (d *DockerService) HealthStatus(name string) (string, error) {
	ctx, cancel := dockerCtx()
	defer cancel()
	info, err := d.cli.ContainerInspect(ctx, name)
	if err != nil {
		return "", nil
	}
	if info.State != nil && info.State.Health != nil {
		hs := info.State.Health.Status
		if hs != "" && hs != "<no value>" {
			return hs, nil
		}
	}
	if info.State != nil {
		switch {
		case info.State.Running:
			return "healthy", nil
		case info.State.Status == "created" || info.State.Status == "restarting":
			return "starting", nil
		default:
			return "unhealthy", nil
		}
	}
	return "", nil
}

// Ice stops the container and removes its image to free disk space.
// The service record is preserved so the container can be redeployed.
func (d *DockerService) Ice(name string) error {
	ctx, cancel := dockerCtx()
	defer cancel()
	info, _ := d.cli.ContainerInspect(ctx, name)
	var img string
	if info.Config != nil {
		img = strings.TrimSpace(info.Config.Image)
	}

	ctx2, cancel2 := dockerCtx()
	defer cancel2()
	if err := d.cli.ContainerStop(ctx2, name, container.StopOptions{}); err != nil {
		return fmt.Errorf("failed to stop container %s: %w", name, err)
	}

	if img != "" {
		ctx3, cancel3 := dockerCtx()
		defer cancel3()
		d.cli.ImageRemove(ctx3, img, image.RemoveOptions{}) //nolint:errcheck
	}
	return nil
}
