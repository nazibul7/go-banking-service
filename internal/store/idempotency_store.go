package store

import (
	"banking-app/internal/model"
	"context"
	"database/sql"
	"encoding/json"
)

type IdempotencyStore struct {
	db *sql.DB
}

func NewIdempotencyStore(db *sql.DB) *IdempotencyStore {
	return &IdempotencyStore{
		db: db,
	}
}

func (s *IdempotencyStore) GetIdempotency(ctx context.Context, idempotencyKey string) (*model.Idempotency, error) {
	query := `SELECT * FROM idempotency_keys WHERE idempotency_key = $1`
	var idempotency model.Idempotency
	err := s.db.QueryRowContext(ctx, query, idempotencyKey).Scan(&idempotency.UserID, &idempotency.IdempotencyKey, &idempotency.Response)
	if err != nil {
		return nil, err
	}
	return &idempotency, nil
}

func (s *IdempotencyStore) InsertIdempotency(ctx context.Context, userID int, idempotencyKey string, response json.RawMessage) error {
	query := `INSERT INTO idempotency_keys(idempotency_key,user_id,response) VALUES ($1,$2,$3)`

	_, err := s.db.ExecContext(ctx, query, userID, response, idempotencyKey)
	if err != nil {
		return err
	}
	return nil
}
