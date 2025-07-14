package models

import "time"

type MagicToken struct {
	Id        string    `db:"id" json:"id"`
	Code      string    `db:"code" json:"code"`
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name,omitempty"`
	Used      bool      `db:"used" json:"used"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
}
