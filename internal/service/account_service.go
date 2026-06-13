package service

import (
	"banking-app/internal/middleware"
	"banking-app/internal/model"
	"context"
	// "database/sql"
	"encoding/json"
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
	CreateAccount(ctx context.Context, balance, userID int) (*model.Account, error)
	GetAccounts(ctx context.Context, userID int) ([]model.Account, error)
	GetAccountByID(ctx context.Context, accountID int) (*model.Account, error)
	UpdateAccount(ctx context.Context, accountID, userID, amount int) error
	DeleteAccount(ctx context.Context, accountID, userID int) error
	TransferTx(ctx context.Context, fromID, toID, amount int) error
}

type IdempotencyStorer interface {
	GetIdempotency(ctx context.Context, idempotencyKey string) (*model.Idempotency, error)
	InsertIdempotency(ctx context.Context, userID int, idempotencyKey string, response json.RawMessage) error
}

type AccountService struct {
	accStore         AccountStorer
	idempotencyStore IdempotencyStorer
}

func NewAccountService(accStore AccountStorer, idempotencyStore IdempotencyStorer) *AccountService {
	return &AccountService{
		accStore:         accStore,
		idempotencyStore: idempotencyStore,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, balance, userID int) (*model.Account, error) {
	if balance <= 0 {
		return nil, errors.New("initial balance can not be negative or zero")
	}

	idempotencyKey, ok := ctx.Value(middleware.IdempotencyKey).(string)
	if !ok || idempotencyKey == "" {
		return nil, errors.New("missing idempotency key")
	}

	existing, err := s.idempotencyStore.GetIdempotency(ctx, idempotencyKey)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if existing != nil {
		var account model.Account
		if err := json.Unmarshal(existing.Response, &account); err != nil {
			return nil, err
		}
		return &account, nil
	}

	account, err := s.accStore.CreateAccount(ctx, balance, userID)
	accountByte, err := json.Marshal(account)
	if err != nil {
		return nil, err
	}

	s.idempotencyStore.InsertIdempotency(ctx, account.UserID, idempotencyKey, accountByte)
	return account, err
}

func (s *AccountService) GetAccounts(ctx context.Context, userID int) ([]model.Account, error) {
	accounts, err := s.accStore.GetAccounts(ctx, userID)
	if err != nil {
		return nil, err
	}

	if userID != accounts[0].UserID {
		return nil, errors.New("unauthorized")
	}

	return accounts, nil
}

func (s *AccountService) Deposit(ctx context.Context, accountID, userID, amount int) error {
	if accountID <= 0 {
		return errors.New("invalid account id")
	}

	if amount <= 0 {
		return errors.New("deposit amount must be greater than zero")
	}

	idempotencyKey, ok := ctx.Value(middleware.IdempotencyKey).(string)
	if !ok || idempotencyKey == "" {
		return errors.New("missing idempotency key")
	}

	existing, err := s.idempotencyStore.GetIdempotency(ctx, idempotencyKey)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if existing != nil {
		var account model.Account
		if err := json.Unmarshal(existing.Response, &account); err != nil {
			return err
		}
		return nil
	}
	acc, err := s.accStore.GetAccountByID(ctx, accountID)
	if err != nil {
		return err
	}
	if acc.UserID != userID {
		return errors.New("forbidden")
	}

	return s.accStore.UpdateAccount(ctx, accountID, userID, amount)
}

func (s *AccountService) Withdraw(ctx context.Context, accountID, userID, amount int) error {
	if accountID <= 0 {
		return errors.New("invalid account id")
	}

	if amount <= 0 {
		return errors.New("withdrawal amount must be greater than zero")
	}

	idempotencyKey, ok := ctx.Value(middleware.IdempotencyKey).(string)
	if !ok || idempotencyKey == "" {
		return nil, errors.New("missing idempotency key")
	}

	existing, err := s.idempotencyStore.GetIdempotency(ctx, idempotencyKey)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if existing != nil {
		var account model.Account
		if err := json.Unmarshal(existing.Response, &account); err != nil {
			return nil, err
		}
		return &account, nil
	}

	acc, err := s.accStore.GetAccountByID(ctx, accountID)
	if err != nil {
		return err
	}
	if acc.UserID != userID {
		return errors.New("forbidden")
	}

	return s.accStore.UpdateAccount(ctx, accountID, userID, -amount)
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

	idempotencyKey, ok := ctx.Value(middleware.IdempotencyKey).(string)
	if !ok || idempotencyKey == "" {
		return nil, errors.New("missing idempotency key")
	}

	existing, err := s.idempotencyStore.GetIdempotency(ctx, idempotencyKey)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if existing != nil {
		var account model.Account
		if err := json.Unmarshal(existing.Response, &account); err != nil {
			return nil, err
		}
		return &account, nil
	}

	sender, err := s.accStore.GetAccountByID(ctx, req.FromID)
	if err != nil {
		return fmt.Errorf("sender: %w", err)
	}

	if sender.UserID != userID {
		return fmt.Errorf("forbidden to send")
	}
	if _, err := s.accStore.GetAccountByID(ctx, req.ToID); err != nil {
		return fmt.Errorf("receiver: %w", err)
	}

	if sender.Balance < req.Amount {
		return fmt.Errorf("insufficient balance: have %d need %d", sender.Balance, req.Amount)
	}

	return s.accStore.TransferTx(ctx, req.FromID, req.ToID, req.Amount)
}

func (s *AccountService) GetAccountByID(ctx context.Context, accountID, userID int) (*model.Account, error) {
	if accountID <= 0 {
		return nil, errors.New("invalid account id")
	}

	account, err := s.accStore.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if userID != account.UserID {
		return nil, errors.New("unauthorized")
	}

	return account, nil
}

func (s *AccountService) DeleteAccount(ctx context.Context, accountID, userID int) error {
	if accountID <= 0 {
		return errors.New("invalid account id")
	}

	idempotencyKey, ok := ctx.Value(middleware.IdempotencyKey).(string)
	if !ok || idempotencyKey == "" {
		return nil, errors.New("missing idempotency key")
	}

	existing, err := s.idempotencyStore.GetIdempotency(ctx, idempotencyKey)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if existing != nil {
		var account model.Account
		if err := json.Unmarshal(existing.Response, &account); err != nil {
			return nil, err
		}
		return &account, nil
	}
	
	return s.accStore.DeleteAccount(ctx, accountID, userID)
}
