// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

// Package tasks models remotely-triggered actions and their results for dployr's distributed task system.
// It defines the Task and Result types, along with the addressing convention ("path:method") for routing
// tasks to appropriate handlers. Execution and orchestration live in the syncer and executor under internal/system.
package tasks
