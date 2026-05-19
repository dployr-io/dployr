// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"os/user"

	"github.com/dployr-io/dployr/pkg/store"
)

// SystemGroupToRole maps Unix group names to dployr roles.
var SystemGroupToRole = map[string]store.Role{
	"dployr-owner":  store.RoleOwner,
	"dployr-admin":  store.RoleAdmin,
	"dployr-dev":    store.RoleDeveloper,
	"dployr-viewer": store.RoleViewer,
}

// highestRoleFromGroups returns the highest dployr role found in a list of
// Unix group names. Returns RoleViewer if no dployr group is present.
func highestRoleFromGroups(groups []string) store.Role {
	role := store.RoleViewer
	level := store.RoleLevel[role]
	for _, g := range groups {
		if r, ok := SystemGroupToRole[g]; ok {
			if store.RoleLevel[r] > level {
				role = r
				level = store.RoleLevel[r]
			}
		}
	}
	return role
}

// GetUserSystemRole returns the highest dployr role for username derived from
// Unix group membership. Defaults to RoleViewer if no dployr group is found.
func GetUserSystemRole(username string) (store.Role, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return store.RoleViewer, err
	}

	gids, err := u.GroupIds()
	if err != nil {
		return store.RoleViewer, err
	}

	names := make([]string, 0, len(gids))
	for _, gid := range gids {
		g, err := user.LookupGroupId(gid)
		if err != nil {
			continue // group may have been deleted from the system
		}
		names = append(names, g.Name)
	}

	return highestRoleFromGroups(names), nil
}

// GetCurrentUserSystemRole returns the current process user's highest dployr role.
func GetCurrentUserSystemRole() (store.Role, error) {
	u, err := user.Current()
	if err != nil {
		return store.RoleViewer, err
	}
	return GetUserSystemRole(u.Username)
}
