package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SigninRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenUser struct {
	ID        int
	UserID    int
	Role      Role
	Email     string
	TokenHash string
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type Claims struct {
	UserID    int       `json:"user_id"`
	Email     string    `json:"email"`
	Role      Role      `json:"role"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
	// doing embedding because we dont have to implement RegisteredClaims methode & It promotes fields like:
	// ExpiresAt, IssuedAt, NotBefore directly into Claims
}
