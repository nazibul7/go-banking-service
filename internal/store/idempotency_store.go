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

<<<<<<< HEAD
func (s *IdempotencyStor) GetIdempotency(ctx context.Context, idempotencyKey string) (*model.Idempotency, error) {
=======
func (s *IdempotencyStore) GetIdempotency(ctx context.Context, idempotencyKey string) (*model.Idempotency, error) {
>>>>>>> 8a529d3 (add getaccounts in store, service, handler so that client could know his accounts)
	query := `SELECT * FROM idempotency_keys WHERE idempotency_key = $1`
	var idempotency model.Idempotency
	err := s.db.QueryRowContext(ctx, query, idempotencyKey).Scan(&idempotency.UserID, &idempotency.IdempotencyKey, &idempotency.Response)
	if err != nil {
		return nil, err
	}
	return &idempotency, nil
}

<<<<<<< HEAD
func (s *IdempotencyStor) InsertIdempotency(ctx context.Context, userID int, idempotencyKey string, response json.RawMessage) error {
	query := `INSERT INTO idempotency_keys(idempotency_key,user_id,response) VALUES ($1,$2,$3)`

	_, err := s.db.ExecContext(ctx, query, userID, response, idempotencyKey)
=======
func (s *IdempotencyStore) InsertIdempotency(ctx context.Context, userID int, idempotencyKey string, response json.RawMessage) error {
	query := `INSERT INTO idempotency_keys(idempotency_key,user_id,response) VALUES ($1,$2,$3)`

	_, err := s.db.ExecContext(ctx, query, idempotencyKey, userID, response)
>>>>>>> 8a529d3 (add getaccounts in store, service, handler so that client could know his accounts)
	if err != nil {
		return err
	}
	return nil
}
