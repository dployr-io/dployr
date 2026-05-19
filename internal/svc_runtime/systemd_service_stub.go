// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package svc_runtime

// SystemdManager is a no-op stub on non-Linux platforms. The real implementation
// (systemd_service.go) uses go-systemd dbus and only compiles on Linux.
type SystemdManager struct {
	DockerService
}
