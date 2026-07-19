package store

import (
	"banking-app/internal/model"
	"context"
	"encoding/json"
	"time"
)

type IdempotencyStore struct{}

func NewIdempotencyStore() *IdempotencyStore {
	return &IdempotencyStore{}
}

func (s *IdempotencyStore) GetIdempotency(ctx context.Context, db DBTX, userID int, idempotencyKey string) (*model.Idempotency, error) {
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
	err := db.QueryRowContext(ctx, query, userID, idempotencyKey).Scan(
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

func (s *IdempotencyStore) InsertIdempotency(ctx context.Context, db DBTX, userID int, idempotencyKey string, statusCode int, response json.RawMessage, expiresAt time.Time) error {
	query := `INSERT INTO idempotency_keys(idempotency_key,user_id,status_code,response, expires_at) VALUES ($1,$2,$3,$4,$5)`

	_, err := db.ExecContext(ctx, query, idempotencyKey, userID, statusCode, response, expiresAt)
	if err != nil {
		return err
	}
	return nil
}
