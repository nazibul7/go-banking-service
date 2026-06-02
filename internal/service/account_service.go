package service

import (
	"banking-app/internal/model"
	"context"
	"errors"
	"fmt"
)

// The interface should be defined where it's used, not where it's implemented — this is the standard Go idiom

// returning *model.Account instead of model.Account because:
//
// 1. nil clearly represents "no result" (e.g. account not found),
//    whereas a zero-value struct like {ID:0, Balance:0} is ambiguous.
//
// 2. avoids copying the struct on every return; important as the struct grows.
//
// 3. allows functions to modify the same object (no accidental copies)

type AccountStorer interface {
	CreateAccount(ctx context.Context, balance int, userID int) (*model.Account, error)
	GetAccount(ctx context.Context, id int) (*model.Account, error)
	UpdateAccount(ctx context.Context, id int, amount int) error
	DeleteAccount(ctx context.Context, id int) error
	TransferTx(ctx context.Context, fromID, toID, amount int) error
}

type AccountService struct {
	store AccountStorer
}

func NewAccountService(store AccountStorer) *AccountService {
	return &AccountService{
		store: store,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, balance, userID int) (*model.Account, error) {
	if balance < 0 {
		return nil, errors.New("initial balance can not be negative")
	}
	return s.store.CreateAccount(ctx, balance, userID)
}

func (s *AccountService) GetAccount(ctx context.Context, accountID, userID int) (*model.Account, error) {
	if accountID <= 0 {
		return nil, errors.New("invalid account id")
	}

	/**
	service layer should NOT know:
	HTTP,middleware, JWT, request context internals
	Service should only know business data:userID,accountID

	That's why didn't used claims from context value

	claims := ctx.Value(middleware.ClaimsKey).(*model.Claims)
	*/

	account, err := s.store.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if userID != account.UserID {
		return nil, errors.New("unauthorized")
	}

	return account, nil
}

func (s *AccountService) Deposit(ctx context.Context, id, amount int) error {
	if id <= 0 {
		return errors.New("invalid account id")
	}

	if amount <= 0 {
		return errors.New("deposit amount must be greater than zero")
	}
	return s.store.UpdateAccount(ctx, id, amount)
}

func (s *AccountService) Withdraw(ctx context.Context, id, amount int) error {
	if id <= 0 {
		return errors.New("invalid account id")
	}

	if amount <= 0 {
		return errors.New("withdrawal amount must be greater than zero")
	}
	return s.store.UpdateAccount(ctx, id, -amount)
}

func (s *AccountService) Transfer(ctx context.Context, req model.TransferRequest, userID int) error {
	if req.Amount <= 0 {
		return errors.New("transfer amount must be greater than zero")
	}
	if req.FromID <= 0 || req.ToID <= 0 {
		return errors.New("invalid account id")
	}
	if req.FromID == req.ToID {
		return errors.New("cannot transfer to same account")
	}

	sender, err := s.store.GetAccount(ctx, req.FromID)
	if err != nil {
		return fmt.Errorf("sender: %w", err)
	}

	if sender.UserID != userID {
		return fmt.Errorf("forbidden to send")
	}
	if _, err := s.store.GetAccount(ctx, req.ToID); err != nil {
		return fmt.Errorf("receiver: %w", err)
	}

	if sender.Balance < req.Amount {
		return fmt.Errorf("insufficient balance: have %d need %d", sender.Balance, req.Amount)
	}

	return s.store.TransferTx(ctx, req.FromID, req.ToID, req.Amount)
}

func (s *AccountService) DeleteAccount(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("invalid account id")
	}
	return s.store.DeleteAccount(ctx, id)
}
