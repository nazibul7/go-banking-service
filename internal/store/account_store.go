package store

import (
	"banking-app/internal/model"
	"context"
	"database/sql"
	"errors"
)

type AccountStore struct{}

func NewAccountStore() *AccountStore {
	return &AccountStore{}
}

func (s *AccountStore) CreateAccount(ctx context.Context, db DBTX, balance, userID int) (*model.Account, error) {
	query := `INSERT INTO accounts (balance, user_id) VALUES ($1,$2) RETURNING id`
	var accountID int
	err := db.QueryRowContext(ctx, query, balance, userID).Scan(&accountID)
	if err != nil {
		return nil, err
	}
	return &model.Account{AccountID: accountID, UserID: userID, Balance: balance}, nil
}

func (s *AccountStore) GetAccounts(ctx context.Context, db DBTX, userID int) ([]model.Account, error) {
	query := `SELECT id, user_id, balance FROM accounts WHERE user_id=$1`
	var accounts []model.Account
	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	/**
		What if I read all data?
		When you do:
	    for rows.Next() {
	      ...
		}

		until completion, many drivers will automatically consume the rest of the result set and make the connection reusable.

		However, the Go documentation still recommends: defer rows.Close()
		because:

		You might return early.
		A scan might fail.
		Future code changes may skip reading all rows.
		It's the standard safe pattern.
	*/
	defer rows.Close()

	for rows.Next() {
		var acc model.Account
		err := rows.Scan(&acc.AccountID, &acc.UserID, &acc.Balance)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (s *AccountStore) GetAccountByID(ctx context.Context, db DBTX, accountID int) (*model.Account, error) {
	query := `SELECT id, user_id, balance FROM accounts WHERE id = $1`
	var account model.Account

	err := db.QueryRowContext(ctx, query, accountID).Scan(&account.AccountID, &account.UserID, &account.Balance)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (s *AccountStore) UpdateAccount(ctx context.Context, db DBTX, accountID, userID int, amount int) (*model.Account, error) {
	var query string
	if amount < 0 {
		query = `UPDATE accounts SET balance = balance + $1 WHERE id = $2 AND user_id = $3 AND balance + $1 >=0
					RETURNING id, user_id, balance`
	} else {
		query = `UPDATE accounts SET balance = balance + $1 WHERE id = $2 AND user_id = $3 RETURNING id, user_id, balance`
	}

	var account model.Account

	err := db.QueryRowContext(ctx, query, amount, accountID, userID).Scan(&account.AccountID, &account.UserID, &account.Balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &account, nil
}

func (s *AccountStore) Transfer(ctx context.Context, db DBTX, fromID, toID, amount int) error {
	var firstID, secondID int
	var firstAmount, secondAmount int

	// consistent lock ordering to prevent deadlock
	if fromID < toID {
		firstID, firstAmount = fromID, -amount
		secondID, secondAmount = toID, amount
	} else {
		firstID, firstAmount = toID, amount
		secondID, secondAmount = fromID, -amount
	}
	var query string
	query = `UPDATE accounts SET balance = balance + $1 WHERE id=$2 AND balance + $1 >=0`

	// deduct from sender
	result, err := db.ExecContext(ctx, query, firstAmount, firstID)
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

	// credit to receiver

	query = `UPDATE accounts SET balance = balance + $1 WHERE id=$2`
	result, err = db.ExecContext(ctx, query, secondAmount, secondID)
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
	return nil
}

func (s *AccountStore) DeleteAccount(ctx context.Context, db DBTX, accountID, userID int) error {
	query := `DELETE FROM accounts WHERE id=$1 AND user_id = $2`

	result, err := db.ExecContext(ctx, query, accountID, userID)
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
