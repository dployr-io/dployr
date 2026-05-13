// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

// Package deploy implements deployment orchestration, handlers, and integration logic
// for managing application deployments in the daemon. It coordinates deployment workflows,
// lifecycle management, and high-level operations, utilizing core types from pkg/core/deploy.
// Concrete deployment execution and internal details are encapsulated within this package.
//
// Build pipeline (build nodes only):
//
//   - BuildImage(name, srcDir, cfg) builds a Docker image from a cloned source
//     directory, pushes it to cfg.RegistryURL, removes the local copy to reclaim
//     disk, and returns the fully-qualified image reference. The image tag is a
//     millisecond Unix timestamp for natural sort order and uniqueness.
//     Requires REGISTRY_URL in config.toml. Set REGISTRY_AUTH to a base64-encoded
//     JSON credential string ({"username":"…","password":"…"}) for authenticated
//     registries such as DigitalOcean Container Registry.
//
//   - imageRef(registryURL, name) constructs the image reference used as the tag
//     for both docker build and docker push.
package deploy
