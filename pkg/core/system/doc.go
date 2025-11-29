// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

// Package system defines the public contract for daemon-level operations in dployr.
//
// It models system health, registration, installation/upgrade, domain requests,
// mode management, and instance lifecycle sync with base over WebSockets. The
// concrete implementation lives in internal/system; consumers should depend on
// the System interface defined here.
package system
