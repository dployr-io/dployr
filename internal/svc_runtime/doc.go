// Package svc_runtime provides cross-platform abstractions for managing system services
// (such as systemd on Linux, launchd on macOS, and NSSM on Windows) used to run and control
// background application processes. It defines core service management interfaces and lifecycle
// operations, with concrete platform-specific implementations located in the internal/svc_runtime package.
package svc_runtime
