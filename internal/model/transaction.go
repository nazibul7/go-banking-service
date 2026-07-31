package model

import "time"

type TransactionType string

const (
	TransactionDeposit  TransactionType = "deposit"
	TransactionWithdraw TransactionType = "withdraw"
	TransactionTransfer TransactionType = "transfer"
)

type TransactionStatus string

const (
	TransactionPending   TransactionStatus = "pending"
	TransactionCompleted TransactionStatus = "completed"
	TransactionFailed    TransactionStatus = "failed"
)

type Transaction struct {
	ID              int
	FromAccountID   *int
	ToAccountID     *int
	Amount          int
	TransactionType TransactionType
	IdempotencyKey  string
	Status          TransactionStatus
	CreatedAt       time.Time
}
