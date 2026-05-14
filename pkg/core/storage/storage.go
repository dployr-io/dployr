// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package storage

import "context"

type MountRequest struct {
	Device     string `json:"device"`
	MountPoint string `json:"mount_point"`
	Token      string `json:"token"`
}

type Mounter interface {
	Mount(ctx context.Context, device, mountPoint string) error
}
