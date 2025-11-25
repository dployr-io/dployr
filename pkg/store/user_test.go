package store

import "testing"

func TestRolePermissions(t *testing.T) {
	tests := []struct {
		current  Role
		required Role
		expected bool
	}{
		{RoleOwner, RoleOwner, true},
		{RoleOwner, RoleAdmin, true},
		{RoleOwner, RoleDeveloper, true},
		{RoleOwner, RoleViewer, true},
		{RoleAdmin, RoleOwner, false},
		{RoleAdmin, RoleAdmin, true},
		{RoleAdmin, RoleDeveloper, true},
		{RoleDeveloper, RoleAdmin, false},
		{RoleDeveloper, RoleDeveloper, true},
		{RoleViewer, RoleDeveloper, false},
		{RoleViewer, RoleViewer, true},
	}

	for _, tt := range tests {
		result := tt.current.IsPermitted(tt.required)
		if result != tt.expected {
			t.Errorf("Role %s.IsPermitted(%s) = %v, want %v",
				tt.current, tt.required, result, tt.expected)
		}
	}
}
