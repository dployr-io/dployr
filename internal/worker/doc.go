// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

// Package worker implements the daemon's background deployment worker, job queue,
// and concurrency management. It coordinates deployment tasks, tracks active jobs,
// and integrates with service orchestration and store layers. All deployment job
// execution logic and internal worker state are encapsulated in this package.
package worker
