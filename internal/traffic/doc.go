// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

// Package traffic analyses Caddy access logs to produce per-service bot-detection
// signals used to distinguish genuine human inactivity from automated keepalive pings.
//
// # Motivation
//
// Hobby-tier services sleep after 30 minutes of genuine inactivity, but external cron
// jobs (e.g. cron-job.org, UptimeRobot) are commonly used to prevent sleep by pinging
// a service on a fixed schedule. A naive last-request timestamp is trivially defeated
// by a single automated pinger.
//
// # Algorithm
//
// Three independent signals are evaluated over a rolling 1-hour window per service.
// A service is considered genuinely idle only when ALL three signals agree:
//
//  1. Subnet diversity — real sites attract visitors from many networks.
//     Threshold: fewer than 3 unique /24 subnets in the window.
//
//  2. Cadence regularity — human visits are aperiodic; cron jobs are metronomic.
//     Measured as the coefficient of variation (stddev/mean) of inter-request gaps.
//     Threshold: CV < 0.2 (highly regular).
//
//  3. Path diversity — humans browse multiple pages; ping bots hit only / or /health.
//     Threshold: fewer than 2 unique request paths in the window.
//
// If any one signal suggests human-like traffic, the service remains awake.
//
// # Data source
//
// Signals are derived from Caddy's per-domain JSON access logs written to
// ~/.dployr/logs/caddy/{domain}.log. No additional instrumentation is required.
// The computed signals are included in the v1.1 node update and evaluated on base.
package traffic
