package store

import "time"

type Remote struct {
	ID         string    `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Repository string    `json:"repository" db:"repository"`
	Branch     string    `json:"branch" db:"branch"`
	Provider   string    `json:"provider" db:"provider"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
