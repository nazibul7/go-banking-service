package store

import (
	"banking-app/internal/model"
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type IdempotencyStore struct {
	db *sql.DB
}

func NewIdempotencyStore(db *sql.DB) *IdempotencyStore {
	return &IdempotencyStore{
		db: db,
	}
}

func (s *IdempotencyStore) GetIdempotency(ctx context.Context, userID int, idempotencyKey string) (*model.Idempotency, error) {
	// Idempotency keys are unique per user, not globally.
	// Using both user_id and idempotency_key ensures we retrieve the correct
	// cached response and prevents cross-user collisions.
	query := `SELECT 
	 id,
	 idempotency_key,
	 user_id,
	 status_code,
	 response,
	 created_at,
	 expires_at
	 FROM idempotency_keys WHERE user_id = $1 AND idempotency_key = $2`
	var idempotency model.Idempotency
	err := s.db.QueryRowContext(ctx, query, userID, idempotencyKey).Scan(
		&idempotency.ID,
		&idempotency.IdempotencyKey,
		&idempotency.UserID,
		&idempotency.StatusCode,
		&idempotency.Response,
		&idempotency.CreatedAt,
		&idempotency.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}
	return &idempotency, nil
}

func (s *IdempotencyStore) InsertIdempotency(ctx context.Context, userID int, idempotencyKey string, statusCode int, response json.RawMessage, expiresAt time.Time) error {
	query := `INSERT INTO idempotency_keys(idempotency_key,user_id,status_code,response, expires_at) VALUES ($1,$2,$3,$4,$5)`

	_, err := s.db.ExecContext(ctx, query, idempotencyKey, userID, statusCode, response, expiresAt)
	if err != nil {
		return err
	}
	return nil
}
