package dto

type CreateAccountRequest struct {
	Balance int `json:"balance"`
}

type BalanceRequest struct {
	Amount int `json:"amount"`
}

type TransferRequest struct {
	FromID int `json:"from_id"`
	ToID   int `json:"to_id"`
	Amount int `json:"amount"`
}
