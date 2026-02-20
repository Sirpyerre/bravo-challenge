package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         Role      `json:"role" db:"role"`
	Country      string    `json:"country" db:"country"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Role string

const (
	RoleUser  Role = "USER"
	RoleAgent Role = "AGENT"
	RoleAdmin Role = "ADMIN"
)

// IsPrivileged reporta si el rol tiene acceso elevado (AGENT o ADMIN).
func (r Role) IsPrivileged() bool {
	return r == RoleAgent || r == RoleAdmin
}
