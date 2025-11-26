// Package system implements the daemon's system service: sync, install/upgrade,
// registration, and health. It owns the WebSocket connection to base, periodic
// pulls and state updates, and local task result handling with deduplication.
// Internal-only; external code should use pkg/core/system.
package system
