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

// GetUserSystemRole returns the highest system role for a specific user
func GetUserSystemRole(username string) (store.Role, error) {
	cmd := exec.Command("id", "-Gn", username)
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

// GetCurrentUserSystemRole returns the current user's highest system role
//
// Example: dployr-admin returns store.RoleAdmin
func GetCurrentUserSystemRole() (store.Role, error) {
	currentUser, err := user.Current()
	if err != nil {
		return store.RoleViewer, err
	}

	return GetUserSystemRole(currentUser.Username)
}
