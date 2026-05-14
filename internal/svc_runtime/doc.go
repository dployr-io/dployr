// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

// Package svc_runtime provides cross-platform abstractions for managing system services
// (such as systemd on Linux, launchd on macOS, and NSSM on Windows) used to run and control
// background application processes. It defines core service management interfaces and lifecycle
// operations, with concrete platform-specific implementations located in the internal/svc_runtime package.

// HealthStatus returns the Docker health check status for a named service.
// Returns "healthy", "unhealthy", "starting", or "" (no health check / not found).
package svc_runtime
