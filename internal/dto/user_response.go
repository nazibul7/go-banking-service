package dto

import "banking-app/internal/model"

type UserResponse struct {
	ID    int        `json:"id"`
	Email string     `json:"email"`
	Role  model.Role `json:"role"`
}