// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

import "testing"

func TestResolveServiceHealth(t *testing.T) {
	tests := []struct {
		name         string
		status       string
		pollerResult string
		exitCode     int
		want         ServiceHealth
	}{
		// running — poller result governs
		{name: "running healthy probe", status: "running", pollerResult: "healthy", exitCode: 0, want: HealthHealthy},
		{name: "running degraded probe", status: "running", pollerResult: "degraded", exitCode: 0, want: HealthDegraded},
		{name: "running no probe yet", status: "running", pollerResult: "", exitCode: 0, want: HealthHealthy},

		// stopped — exit code governs, poller is irrelevant
		{name: "stopped clean exit", status: "stopped", pollerResult: "degraded", exitCode: 0, want: HealthHealthy},
		{name: "stopped crash exit 1", status: "stopped", pollerResult: "healthy", exitCode: 1, want: HealthDegraded},
		{name: "stopped crash exit 137", status: "stopped", pollerResult: "", exitCode: 137, want: HealthDegraded},
		{name: "stopped crash exit 255", status: "stopped", pollerResult: "", exitCode: 255, want: HealthDegraded},

		// starting — always healthy (probe not yet meaningful)
		{name: "starting", status: "starting", pollerResult: "", exitCode: 0, want: HealthHealthy},
		{name: "starting with degraded poller", status: "starting", pollerResult: "degraded", exitCode: 1, want: HealthHealthy},

		// unknown status — treated as healthy (defensive)
		{name: "unknown status", status: "", pollerResult: "", exitCode: 1, want: HealthHealthy},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveServiceHealth(tc.status, tc.pollerResult, tc.exitCode)
			if got != tc.want {
				t.Errorf("resolveServiceHealth(%q, %q, %d) = %q; want %q",
					tc.status, tc.pollerResult, tc.exitCode, got, tc.want)
			}
		})
	}
}
