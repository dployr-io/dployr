// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

// Package store defines data models and interfaces for persistence in dployr.
//
// It models users, deployments, services, instances, and task results, and provides
// contract interfaces for interacting with persistent state. All concrete persistence
// implementations, including SQLite-backed stores, reside in the internal/store package.
package store
