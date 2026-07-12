package dto

type CreateAccountResponse struct {
	AccountID int    `json:"account_id"`
	Balance   int    `json:"balance"`
	Message   string `json:"message"`
}

type AccountResponse struct {
	AccountID int `json:"account_id"`
	Balance   int `json:"balance"`
}

type BalanceResponse struct {
	AccountID int    `json:"account_id"`
	Balance   int    `json:"balance"`
	Amount    int    `json:"amount"`
	Message   string `json:"message"`
}

type TransferResponse struct {
	FromID  int    `json:"from_id"`
	ToID    int    `json:"to_id"`
	Amount  int    `json:"amount"`
	Message string `json:"message"`
}

type DeleteAccountResponse struct {
	Message string `json:"message"`
}