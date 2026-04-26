package model

type Account struct {
	ID      int `json:"id"`
	Balance int `json:"balance"`
}

type CreateAccountRequest struct {
	Balance int `json:"balance"`
}

type AmountRequest struct {
	Amount int `json:"amount"`
}
