package store

import (
	"banking-app/internal/model"
	"context"
	"database/sql"
)

type AccountStore struct {
	db *sql.DB
}

func NewAccountStore(db *sql.DB) *AccountStore {
	return &AccountStore{db: db}
}

func (s *AccountStore) CreateAccount(ctx context.Context, balance int) (*model.Account, error) {
	query := `INSERT INTO accounts (balance) VALUES ($1) RETURNING id`
	var id int
	err := s.db.QueryRowContext(ctx, query, balance).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &model.Account{ID: id, Balance: balance}, nil
}

func (s *AccountStore) GetAccount(ctx context.Context, id int) (*model.Account, error) {
	query := `SELECT * FROM accounts WHERE id=$1`
	var acc model.Account
	err := s.db.QueryRowContext(ctx, query, id).Scan(&acc.ID, &acc.Balance)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (s *AccountStore) UpdateAccount(ctx context.Context, id int, amount int) error {
	var query string
	if amount < 0 {
		query = `UPDATE accounts SET balance = balance + $1 WHERE id = $2 AND balance + $1 >=0`
	} else {
		query = `UPDATE accounts SET balance = balance + $1 WHERE id = $2`
	}

	result, err := s.db.ExecContext(ctx, query, amount, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *AccountStore) TransferTx(
	ctx context.Context,
	fromID,
	toID,
	amount int,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// deterministic locking order only
	firstID := fromID
	secondID := toID

	if fromID > toID {
		firstID = toID
		secondID = fromID
	}

	// lock smaller ID first
	var id int
	query := `SELECT id FROM accounts WHERE id=$1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, query, firstID).Scan(&id)
	if err != nil {
		return err
	}

	// lock larger ID second
	err = tx.QueryRowContext(ctx, query, secondID).Scan(&id)
	if err != nil {
		return err
	}

	// actual business logic starts here

	// deduct from sender
	query = `
		UPDATE accounts
		SET balance = balance - $1
		WHERE id=$2 AND balance >= $1
	`

	result, err := tx.ExecContext(ctx, query, amount, fromID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	// add to receiver
	query = `
		UPDATE accounts
		SET balance = balance + $1
		WHERE id=$2
	`

	result, err = tx.ExecContext(ctx, query, amount, toID)
	if err != nil {
		return err
	}

	rows, err = result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}

func (s *AccountStore) DeleteAccount(ctx context.Context, id int) error {
	query := `DELETE FROM accounts WHERE id=$1`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
