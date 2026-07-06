package model

import "time"

type RefreshToken struct {
	ID        int
	UserID    int
	Role      Role
	Email     string
	TokenHash string
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}