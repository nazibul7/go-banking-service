package dto

import "time"

type TransactionResponse struct {
	ID              int       `json:"id"`
	FromAccountID   *int      `json:"from_account_id,omitempty"`
	ToAccountID     *int      `json:"to_account_id,omitempty"`
	Amount          int       `json:"amount"`
	TransactionType string    `json:"transaction_type"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}
