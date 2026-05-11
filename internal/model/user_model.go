package model

import "time"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // json:"-" prevents password hash from leaking in API responses.
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}
