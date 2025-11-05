package auth

import (
	"dployr/pkg/store"
	"os/exec"
	"os/user"
	"strings"
)

// System group to role mapping
var SystemGroupToRole = map[string]store.Role{
	"dployr-owner":  store.RoleOwner,
	"dployr-admin":  store.RoleAdmin,
	"dployr-dev":    store.RoleDeveloper,
	"dployr-viewer": store.RoleViewer,
}

// GetCurrentUserSystemRole returns the current user's highest system role
func GetCurrentUserSystemRole() (store.Role, error) {
	currentUser, err := user.Current()
	if err != nil {
		return store.RoleViewer, err
	}

	// Get user's groups
	cmd := exec.Command("groups", currentUser.Username)
	output, err := cmd.Output()
	if err != nil {
		return store.RoleViewer, err
	}

	userGroups := strings.Fields(string(output))

	// Find highest role from group membership
	highestRole := store.RoleViewer
	highestLevel := store.RoleLevel[highestRole]

	for _, group := range userGroups {
		if role, exists := SystemGroupToRole[group]; exists {
			if store.RoleLevel[role] > highestLevel {
				highestRole = role
				highestLevel = store.RoleLevel[role]
			}
		}
	}

	return highestRole, nil
}
