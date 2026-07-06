package dto

import "banking-app/internal/model"

type UserResponse struct {
	ID    int        `json:"id"`
	Email string     `json:"email"`
	Role  model.Role `json:"role"`
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}
