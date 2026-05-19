// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"testing"

	"github.com/dployr-io/dployr/pkg/store"
)

func TestHighestRoleFromGroups_OwnerWins(t *testing.T) {
	groups := []string{"staff", "dployr-viewer", "dployr-dev", "dployr-owner"}
	if got := highestRoleFromGroups(groups); got != store.RoleOwner {
		t.Errorf("expected RoleOwner, got %s", got)
	}
}

func TestHighestRoleFromGroups_AdminBeatsViewer(t *testing.T) {
	got := highestRoleFromGroups([]string{"dployr-viewer", "dployr-admin"})
	if got != store.RoleAdmin {
		t.Errorf("expected RoleAdmin, got %s", got)
	}
}

func TestHighestRoleFromGroups_DeveloperOnly(t *testing.T) {
	got := highestRoleFromGroups([]string{"docker", "dployr-dev"})
	if got != store.RoleDeveloper {
		t.Errorf("expected RoleDeveloper, got %s", got)
	}
}

func TestHighestRoleFromGroups_NoDployrGroups(t *testing.T) {
	got := highestRoleFromGroups([]string{"staff", "docker", "wheel"})
	if got != store.RoleViewer {
		t.Errorf("expected RoleViewer (default) when no dployr groups, got %s", got)
	}
}

func TestHighestRoleFromGroups_Empty(t *testing.T) {
	if got := highestRoleFromGroups(nil); got != store.RoleViewer {
		t.Errorf("expected RoleViewer for nil groups, got %s", got)
	}
	if got := highestRoleFromGroups([]string{}); got != store.RoleViewer {
		t.Errorf("expected RoleViewer for empty groups, got %s", got)
	}
}

func TestHighestRoleFromGroups_UnknownGroupsIgnored(t *testing.T) {
	got := highestRoleFromGroups([]string{"dployr-unknown", "not-a-role", "dployr-dev"})
	if got != store.RoleDeveloper {
		t.Errorf("expected RoleDeveloper ignoring unknown groups, got %s", got)
	}
}

func TestSystemGroupToRole_AllRolesCovered(t *testing.T) {
	required := []store.Role{store.RoleOwner, store.RoleAdmin, store.RoleDeveloper, store.RoleViewer}
	present := map[store.Role]bool{}
	for _, r := range SystemGroupToRole {
		present[r] = true
	}
	for _, r := range required {
		if !present[r] {
			t.Errorf("SystemGroupToRole is missing an entry for role %q", r)
		}
	}
}

func TestSystemGroupToRole_GroupNamesHavePrefix(t *testing.T) {
	for group := range SystemGroupToRole {
		if len(group) < 7 || group[:7] != "dployr-" {
			t.Errorf("group %q does not follow dployr- naming convention", group)
		}
	}
}
