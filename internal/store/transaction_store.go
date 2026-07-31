package store

import (
	"banking-app/internal/model"
	"context"
	"database/sql"
)

type TransactionStore struct{}

func NewTransactionStore() *TransactionStore {
	return &TransactionStore{}
}

func (t *TransactionStore) CreateTransaction(
	ctx context.Context,
	db DBTX,
	fromAccountID,
	toAccountID *int,
	amount int,
	idempotencyKey string,
	transactionType model.TransactionType,
	transactionStatus model.TransactionStatus,
) error {

	query := `
		INSERT INTO transactions (
			from_account_id,
			to_account_id,
			amount,
			transaction_type,
			status,
			idempotency_key
		)
		VALUES ($1,$2,$3,$4,$5,$6)
	`

	_, err := db.ExecContext(
		ctx,
		query,
		fromAccountID,
		toAccountID,
		amount,
		transactionType,
		transactionStatus,
		idempotencyKey,
	)

	return err
}

func (t *TransactionStore) GetTransactions(
	ctx context.Context,
	db DBTX,
	userID int,
) ([]model.Transaction, error) {

	query := `
	SELECT
		t.id,
		t.from_account_id,
		t.to_account_id,
		t.amount,
		t.transaction_type,
		t.status,
		t.idempotency_key,
		t.created_at
	FROM transactions t
	JOIN accounts a
		ON a.id = t.from_account_id
		OR a.id = t.to_account_id
	WHERE a.user_id = $1
	ORDER BY t.created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []model.Transaction

	for rows.Next() {
		var tx model.Transaction

		var fromID sql.NullInt64
		var toID sql.NullInt64

		err := rows.Scan(
			&tx.ID,
			&fromID,
			&toID,
			&tx.Amount,
			&tx.TransactionType,
			&tx.Status,
			&tx.IdempotencyKey,
			&tx.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if fromID.Valid {
			v := int(fromID.Int64)
			tx.FromAccountID = &v
		}

		if toID.Valid {
			v := int(toID.Int64)
			tx.ToAccountID = &v
		}

		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (t *TransactionStore) GetTransactionByID(
	ctx context.Context,
	db DBTX,
	transactionID int,
	userID int,
) (*model.Transaction, error) {
	query := `
	SELECT
		t.id,
		t.from_account_id,
		t.to_account_id,
		t.amount,
		t.transaction_type,
		t.status,
		t.idempotency_key,
		t.created_at
	FROM transactions t
	JOIN accounts a
		ON a.id = t.from_account_id
		OR a.id = t.to_account_id
	WHERE t.id = $1
	  AND a.user_id = $2
	`

	var tx model.Transaction
	var fromID sql.NullInt64
	var toID sql.NullInt64

	err := db.QueryRowContext(
		ctx,
		query,
		transactionID,
		userID,
	).Scan(
		&tx.ID,
		&fromID,
		&toID,
		&tx.Amount,
		&tx.TransactionType,
		&tx.Status,
		&tx.IdempotencyKey,
		&tx.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if fromID.Valid {
		v := int(fromID.Int64)
		tx.FromAccountID = &v
	}

	if toID.Valid {
		v := int(toID.Int64)
		tx.ToAccountID = &v
	}

	return &tx, nil
}

func (t *TransactionStore) GetAccountTransactions(
	ctx context.Context,
	db DBTX,
	accountID int,
	userID int,
) ([]model.Transaction, error) {

	query := `
	SELECT
		t.id,
		t.from_account_id,
		t.to_account_id,
		t.amount,
		t.transaction_type,
		t.status,
		t.idempotency_key,
		t.created_at
	FROM transactions t
	JOIN accounts a
		ON a.id = $1
	WHERE
		(t.from_account_id = $1 OR t.to_account_id = $1)
		AND a.user_id = $2
	ORDER BY t.created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, accountID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []model.Transaction

	for rows.Next() {
		var tx model.Transaction

		// PostgreSQL columns from_account_id and to_account_id are nullable.
		//
		// We cannot scan a nullable SQL column directly into:
		//
		//	int   -> fails when database returns NULL
		//	*int  -> database/sql does not populate *int from SQL NULL
		//
		// Therefore we first scan into sql.NullInt64, which can represent both:
		//
		//	Valid=true  -> column contains an integer
		//	Valid=false -> column is NULL
		//
		// After scanning, we convert sql.NullInt64 into *int for our model.
		//
		// Summary:
		//
		//	Go type                DB = 5      DB = NULL
		//	-------------------------------------------------
		//	int                    ✓           ✗ Scan error
		//	*int (direct Scan)     ✗           ✗ Unsupported
		//	sql.NullInt64          ✓           ✓
		//
		// This allows the API to return:
		//   - nil for deposit/withdraw missing account IDs
		//   - actual account IDs for transfer transactions.
		var fromID sql.NullInt64
		var toID sql.NullInt64

		err := rows.Scan(
			&tx.ID,
			&fromID,
			&toID,
			&tx.Amount,
			&tx.TransactionType,
			&tx.Status,
			&tx.IdempotencyKey,
			&tx.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Convert sql.NullInt64 -> *int.
		//
		// If Valid is false, the database value was NULL,
		// so the pointer remains nil.
		//
		// If Valid is true, create an int and store its address.
		// The model now correctly represents nullable account IDs.
		if fromID.Valid {
			v := int(fromID.Int64)
			tx.FromAccountID = &v
		}

		if toID.Valid {
			v := int(toID.Int64)
			tx.ToAccountID = &v
		}

		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
