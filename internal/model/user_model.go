package model

import "time"

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // json:"-" prevents password hash from leaking in API responses.
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}
