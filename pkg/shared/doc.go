// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

// Package shared provides configuration, logging, error helpers, and constants
// used across dployrd. Shared, cross-cutting utilities live here to avoid
// duplication in other packages, including config for sync intervals and
// task deduplication TTLs.
//
// Config fields related to container builds:
//
//   - RegistryURL:  container registry to push images to
//     (e.g. "registry.digitalocean.com/my-registry" or "localhost:5000").
//     Set via REGISTRY_URL in config.toml.
//
//   - RegistryAuth: base64-encoded JSON credentials for docker login
//     ({"username":"…","password":"…"}). Set via REGISTRY_AUTH in config.toml.
//     Leave empty for unauthenticated or credential-helper-backed registries.
package shared
