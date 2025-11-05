package store

import (
	"context"
	"time"
)

type Role string

// owner > admin > developer > viewer
const (
	RoleOwner     Role = "owner"     // uninstall dployr, manage admins
	RoleAdmin     Role = "admin"     // delete infrastructure, secrets, users, proxies; manage deployr versions; shell access
	RoleDeveloper Role = "developer" // deploy apps, view logs, view events, view resource graph
	RoleViewer    Role = "viewer"    // view services
)

var RoleLevel = map[Role]int{
	RoleOwner:     3,
	RoleAdmin:     2,
	RoleDeveloper: 1,
	RoleViewer:    0,
}

func (r Role) IsPermitted(required Role) bool {
	return RoleLevel[r] >= RoleLevel[required]
}

type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Role      Role      `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type UserStore interface {
	FindOrCreateUser(email string, role Role) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUserRole(ctx context.Context, email string, role Role) error
	HasOwner() (bool, error) // Returns true if owner exists
	GetRole(ctx context.Context, email string) (Role, error)
}
