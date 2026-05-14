// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

// Package store defines data models and interfaces for persistence in dployr.
//
// It models users, deployments, services, instances, and task results, and provides
// contract interfaces for interacting with persistent state. All concrete persistence
// implementations, including SQLite-backed stores, reside in the internal/store package.
//
// Unit conventions for ResourceLimits fields:
//
//	Memory  — megabytes (MB).  e.g. 64 = 64 MB, 512 = 512 MB
//	CPU     — millicores.      e.g. 100 = 0.1 vCPU, 1000 = 1 vCPU
//	Storage — gigabytes (GB).  e.g. 5 = 5 GB, 25 = 25 GB
//
// Token is a short-lived credential injected by the control plane at dispatch
// time. It is used to authenticate git clone for private repositories and is
// never persisted to long-term storage.

package store
