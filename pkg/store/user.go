package store

import (
	"context"
	"time"
)

type Role string

// TODO: Proper roles description
// Owner > Admin > Developer > Viewer
const (
	RoleOwner    Role = "owner"     // Max priviledges 
	RoleAdmin    Role = "admin"     // Manage deployments, teams, etc.
	RoleDeveloper Role = "developer" // Deploy new services, update 
	RoleViewer   Role = "viewer"    // Read-only access
)

var RoleLevel = map[Role]int{
	RoleOwner:      3,
	RoleAdmin:      2,
	RoleDeveloper:  1,
	RoleViewer:     0,
}

func (r Role) IsPermitted(required Role) bool {
	return RoleLevel[r] >= RoleLevel[required]
}


type User struct {
	ID            string         `json:"id" db:"id"`
	Name          string         `json:"name" db:"name"`
	Email         string         `json:"email" db:"email"`
	Role 		  Role           `josn:"role" db:"role"`
	Password      string         `json:"-" db:"password"`
	MagicToken        *string    `db:"magic_token,omitempty" json:"-"`
    MagicTokenExpiry  *time.Time `db:"magic_token_expiry,omitempty" json:"-"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
}

type UserStore interface {
	FindOrCreateUser(email string) (*User, error)
	SaveMagicToken(email, token string, expiry time.Time) error
	ValidateMagicToken(token string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}
