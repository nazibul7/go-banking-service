package store

import (
	"banking-app/internal/model"
	"context"
	"database/sql"
	"time"
)

type TxStore struct {
	db *sql.DB
}

func NewTxStore(db *sql.DB) *TxStore {
	return &TxStore{db: db}
}

func (s *TxStore) RegisterTx(ctx context.Context, email, passwordHash, newHashRefreshToken string, expires_at time.Time) (*model.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	var user model.User

	query := `INSERT INTO users (email,password_hash) VALUES ($1, $2) RETURNING id, email, role, created_at`
	err = tx.QueryRowContext(ctx, query, email, passwordHash).Scan(&user.ID, &user.Email, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	query = `INSERT INTO refresh_tokens (user_id,token_hash,expires_at) VALUES ($1, $2,$3)`
	_, err = tx.ExecContext(ctx, query, user.ID, newHashRefreshToken, expires_at)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *TxStore) RotateTokenTx(ctx context.Context, userID int, oldhashToken, newHashToken string, expiresAt time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	revokeQuery := `UPDATE refresh_tokens
	SET revoked = TRUE
	WHERE token_hash = $1`

	_, err = tx.ExecContext(ctx, revokeQuery, oldhashToken)
	if err != nil {
		return err
	}

	insertQuery := `INSERT INTO refresh_tokens(user_id,token_hash, expires_at) VALUES($1, $2, $3)`
	_, err = tx.ExecContext(ctx, insertQuery, userID, newHashToken, expiresAt)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
