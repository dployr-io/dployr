// Package system defines the public contract for daemon-level operations in dployr.
//
// It models system health, registration, installation/upgrade, domain requests,
// and mode management. Consumers should depend on the System interface defined
// here; the concrete implementation lives in internal/system.
package system
