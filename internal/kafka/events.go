package kafka

import (
	"banking-app/internal/model"
	"time"
)

type AccountCreatedEvent struct {
	AccountID int       `json:"account_id"`
	UserID    int       `json:"user_id"`
	Balance   int       `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

type DepositCompletedEvent struct {
	TransactionID   int                   `json:"transaction_id"`
	AccountID       int                   `json:"account_id"`
	TransactionType model.TransactionType `json:"transaction_type"`
	Amount          int                   `json:"amount"`
	OccurredAt      time.Time             `json:"occurred_at"`
}

type WithdrawCompletedEvent struct {
	TransactionID   int                   `json:"transaction_id"`
	AccountID       int                   `json:"account_id"`
	TransactionType model.TransactionType `json:"transaction_type"`
	Amount          int                   `json:"amount"`
	OccurredAt      time.Time             `json:"occurred_at"`
}

type TransferCompletedEvent struct {
	TransactionID   int                   `json:"transaction_id"`
	FromAccountID   int                   `json:"from_account_id"`
	ToAccountID     int                   `json:"to_account_id"`
	TransactionType model.TransactionType `json:"transaction_type"`
	Amount          int                   `json:"amount"`
	OccurredAt      time.Time             `json:"occurred_at"`
}
