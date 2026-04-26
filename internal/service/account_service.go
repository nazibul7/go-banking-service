package service

import (
	"banking-app/internal/model"
	"context"
	"errors"
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
	CreateAccount(ctx context.Context, balance int) (*model.Account, error)
	GetAccount(ctx context.Context, id int) (*model.Account, error)
	UpdateAccount(ctx context.Context, id int, amount int) error
	DeleteAccount(ctx context.Context, id int) error
}

type AccountService struct {
	store AccountStorer
}

func NewAccountService(store AccountStorer) *AccountService {
	return &AccountService{
		store: store,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, balance int) (*model.Account, error) {
	if balance < 0 {
		return nil, errors.New("initial balance can not be negative")
	}
	return s.store.CreateAccount(ctx, balance)
}

func (s *AccountService) GetAccount(ctx context.Context, id int) (*model.Account, error) {
	if id <= 0 {
		return nil, errors.New("invalid account id")
	}
	return s.store.GetAccount(ctx, id)
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

func (s *AccountService) DeleteAccount(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("invalid account id")
	}
	return s.store.DeleteAccount(ctx, id)
}
