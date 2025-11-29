// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package auth

import "testing"

func TestSystemGroupToRole(t *testing.T) {
	expectedMappings := map[string]string{
		"dployr-owner":  "owner",
		"dployr-admin":  "admin",
		"dployr-dev":    "developer",
		"dployr-viewer": "viewer",
	}

	for group, expectedRole := range expectedMappings {
		role, exists := SystemGroupToRole[group]
		if !exists {
			t.Errorf("Group %q not found in SystemGroupToRole", group)
			continue
		}
		if string(role) != expectedRole {
			t.Errorf("SystemGroupToRole[%q] = %q, want %q", group, role, expectedRole)
		}
	}
}
