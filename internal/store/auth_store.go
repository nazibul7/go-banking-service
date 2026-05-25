package store

import (
	"banking-app/internal/model"
	"context"
	"database/sql"
)

type AuthStore struct {
	db *sql.DB
}

func NewAuthStore(db *sql.DB) *AuthStore {
	return &AuthStore{db: db}
}

func (s *AuthStore) CreateUser(ctx context.Context, email, passwordHash string) (*model.User, error) {
	query := `INSERT INTO users (email,password_hash) VALUES ($1, $2) RETURNING id, email, role, created_at`

	var user model.User

	err := s.db.QueryRowContext(ctx, query, email, passwordHash).Scan(
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

func (s *AuthStore) SigninUser(ctx context.Context, email, passwordHash string) (*model.User, error) {
	query := `INSERT INTO users (email,password_hash) VALUES ($1, $2) RETURNING id, email, role, created_at`

	var user model.User

	err := s.db.QueryRowContext(ctx, query, email, passwordHash).Scan(
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

func (s *AuthStore) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, email, password_hash, role, created_at FROM users WHERE email=$1`

	var user model.User
	err := s.db.QueryRowContext(ctx, query, email).Scan(
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
