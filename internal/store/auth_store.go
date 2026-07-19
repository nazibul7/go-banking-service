package store

import (
	"banking-app/internal/model"
	"context"
)

type AuthStore struct{}

func NewAuthStore() *AuthStore {
	return &AuthStore{}
}

func (s *AuthStore) CreateUser(ctx context.Context, db DBTX, email, passwordHash string) (*model.User, error) {
	query := `INSERT INTO users (email,password_hash) VALUES ($1, $2) RETURNING id, email, role, created_at`

	var user model.User

	err := db.QueryRowContext(ctx, query, email, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthStore) SigninUser(ctx context.Context, db DBTX, email, passwordHash string) (*model.User, error) {
	query := `INSERT INTO users (email,password_hash) VALUES ($1, $2) RETURNING id, email, role, created_at`

	var user model.User

	err := db.QueryRowContext(ctx, query, email, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthStore) GetUserByEmail(ctx context.Context, db DBTX, email string) (*model.User, error) {
	query := `SELECT id, email, password_hash, role, created_at FROM users WHERE email=$1`

	var user model.User
	err := db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &user, nil
}
