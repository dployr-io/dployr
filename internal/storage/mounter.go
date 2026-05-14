// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dployr-io/dployr/pkg/shared"
)

type DefaultMounter struct {
	logger *shared.Logger
}

func NewMounter(l *shared.Logger) *DefaultMounter {
	return &DefaultMounter{logger: l}
}

func (m *DefaultMounter) Mount(ctx context.Context, device, mountPoint string) error {
	// Ensure mount point directory exists.
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return fmt.Errorf("failed to create mount point %s: %w", mountPoint, err)
	}

	// Check whether the device already has a filesystem; format if not.
	if !m.hasFilesystem(ctx, device) {
		m.logger.Info("storage: formatting device", "device", device)
		mkfs := exec.CommandContext(ctx, "mkfs.ext4", "-F", device)
		if out, err := mkfs.CombinedOutput(); err != nil {
			return fmt.Errorf("mkfs.ext4 failed: %w: %s", err, string(out))
		}
	}

	// Mount the device.
	m.logger.Info("storage: mounting device", "device", device, "mount_point", mountPoint)
	mnt := exec.CommandContext(ctx, "mount", device, mountPoint)
	if out, err := mnt.CombinedOutput(); err != nil {
		// Already mounted is acceptable — idempotent.
		if strings.Contains(string(out), "already mounted") {
			m.logger.Info("storage: device already mounted", "device", device)
			return nil
		}
		return fmt.Errorf("mount failed: %w: %s", err, string(out))
	}

	m.logger.Info("storage: mount successful", "device", device, "mount_point", mountPoint)
	return nil
}

// hasFilesystem returns true when blkid can identify a filesystem on the device.
func (m *DefaultMounter) hasFilesystem(ctx context.Context, device string) bool {
	out, err := exec.CommandContext(ctx, "blkid", "-s", "TYPE", "-o", "value", device).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}
