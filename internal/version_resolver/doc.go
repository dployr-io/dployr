// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

// Package version_resolver maps a runtime name and user-supplied version string
// to a Docker image reference.
//
// Version inputs are intentionally loose — a major ("3"), a minor ("3.12"),
// or a full patch pin ("3.12.7") are all valid. Missing precision is filled in
// by querying endoflife.date, which picks the latest non-EOL release that
// matches the prefix.
//
// Responses are cached per product for 24 hours. A failed refresh serves the
// previous data so a transient outage to endoflife.date does not block builds.
//
// Runtimes with multi-stage Dockerfiles (nodejs, java) expose separate builder
// and runner images through [Resolution]. Single-stage runtimes return the same
// image for both.
//
// Two inputs are always rejected: strings that do not parse as a version prefix
// (letters, extra dots, wildcards), and versions whose end-of-life date has
// already passed.
package version_resolver
