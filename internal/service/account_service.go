package service

import (
	"banking-app/internal/dto"
	"banking-app/internal/middleware"
	"banking-app/internal/model"
	"banking-app/internal/store"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
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
	CreateAccount(ctx context.Context, db store.DBTX, balance, userID int) (*model.Account, error)
	GetAccounts(ctx context.Context, db store.DBTX, userID int) ([]model.Account, error)
	GetAccountByID(ctx context.Context, db store.DBTX, accountID int) (*model.Account, error)
	UpdateAccount(ctx context.Context, db store.DBTX, accountID, userID, amount int) (*model.Account, error)
	DeleteAccount(ctx context.Context, db store.DBTX, accountID, userID int) error
	Transfer(ctx context.Context, db store.DBTX, fromID, toID, amount int) error
}

type IdempotencyStorer interface {
	GetIdempotency(ctx context.Context, db store.DBTX, userID int, idempotencyKey string) (*model.Idempotency, error)
	InsertIdempotency(ctx context.Context, db store.DBTX, userID int, idempotencyKey string, statusCode int, response json.RawMessage, expiresAt time.Time) error
}

type TransactionStorer interface {
	CreateTransaction(ctx context.Context, db store.DBTX, fromAccountID, toAccountID *int, amount int, idempotencyKey string, transactionType model.TransactionType, transactionStatus model.TransactionStatus) error
}

type AccountService struct {
	db               *sql.DB
	accStore         AccountStorer
	idempotencyStore IdempotencyStorer
	transactionStore TransactionStorer
}

func NewAccountService(db *sql.DB, accStore AccountStorer, idempotencyStore IdempotencyStorer, transactionStore TransactionStorer) *AccountService {
	return &AccountService{
		db:               db,
		accStore:         accStore,
		idempotencyStore: idempotencyStore,
		transactionStore: transactionStore,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, balance, userID int) (*dto.CreateAccountResponse, error) {
	if balance <= 0 {
		return nil, errors.New("initial balance can not be negative or zero")
	}

	idempotencyKey, ok := ctx.Value(middleware.IdempotencyKey).(string)
	if !ok || idempotencyKey == "" {
		return nil, errors.New("missing idempotency key")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	existing, err := s.idempotencyStore.GetIdempotency(ctx, tx, userID, idempotencyKey)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if existing != nil {
		var account dto.CreateAccountResponse
		if err := json.Unmarshal(existing.Response, &account); err != nil {
			return nil, err
		}
		return &account, nil
	}

	account, err := s.accStore.CreateAccount(ctx, tx, balance, userID)
	if err != nil {
		return nil, err
	}

	var response *dto.CreateAccountResponse

	response = &dto.CreateAccountResponse{
		AccountID: account.AccountID,
		Balance:   account.Balance,
		Message:   "Account created successfully",
	}

	responseByte, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	if err := s.idempotencyStore.InsertIdempotency(ctx, tx, account.UserID, idempotencyKey, http.StatusCreated, responseByte, expiresAt); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return response, nil
}

func (s *AccountService) GetAccounts(ctx context.Context, userID int) ([]dto.AccountResponse, error) {
	accounts, err := s.accStore.GetAccounts(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.AccountResponse, 0, len(accounts))

	for _, acc := range accounts {
		responses = append(responses, dto.AccountResponse{
			AccountID: acc.AccountID,
			Balance:   acc.Balance,
		})
	}
	return responses, nil
}

func (s *AccountService) Deposit(ctx context.Context, accountID, userID, amount int) (*dto.BalanceResponse, error) {
	if accountID <= 0 {
		return nil, errors.New("invalid account id")
	}

	if amount <= 0 {
		return nil, errors.New("deposit amount must be greater than zero")
	}

	idempotencyKey, ok := ctx.Value(middleware.IdempotencyKey).(string)
	if !ok || idempotencyKey == "" {
		return nil, errors.New("missing idempotency key")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	existing, err := s.idempotencyStore.GetIdempotency(ctx, tx, userID, idempotencyKey)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if existing != nil {
		var response dto.BalanceResponse
		if err := json.Unmarshal(existing.Response, &response); err != nil {
			return nil, err
		}
		return &response, nil
	}
	acc, err := s.accStore.GetAccountByID(ctx, tx, accountID)
	if err != nil {
		return nil, err
	}
	if acc.UserID != userID {
		return nil, errors.New("forbidden")
	}

	account, err := s.accStore.UpdateAccount(ctx, tx, accountID, userID, amount)
	if err != nil {
		return nil, err
	}

	if err := s.transactionStore.CreateTransaction(ctx, tx, nil, &accountID, amount, idempotencyKey, model.TransactionDeposit, model.TransactionCompleted); err != nil {
		return nil, err
	}

	response := &dto.BalanceResponse{
		AccountID: account.AccountID,
		Balance:   account.Balance,
		Amount:    amount,
		Message:   "Deposit completed successfully",
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	if err := s.idempotencyStore.InsertIdempotency(
		ctx,
		tx,
		userID,
		idempotencyKey,
		http.StatusOK,
		responseBytes,
		expiresAt,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return response, nil
}

func (s *AccountService) Withdraw(ctx context.Context, accountID, userID, amount int) (*dto.BalanceResponse, error) {
	if accountID <= 0 {
		return nil, errors.New("invalid account id")
	}

	if amount <= 0 {
		return nil, errors.New("withdrawal amount must be greater than zero")
	}

	idempotencyKey, ok := ctx.Value(middleware.IdempotencyKey).(string)
	if !ok || idempotencyKey == "" {
		return nil, errors.New("missing idempotency key")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	existing, err := s.idempotencyStore.GetIdempotency(ctx, tx, userID, idempotencyKey)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if existing != nil {
		var response dto.BalanceResponse
		if err := json.Unmarshal(existing.Response, &response); err != nil {
			return nil, err
		}
		return &response, nil
	}

	acc, err := s.accStore.GetAccountByID(ctx, tx, accountID)
	if err != nil {
		return nil, err
	}
	if acc.UserID != userID {
		return nil, errors.New("forbidden")
	}

	account, err := s.accStore.UpdateAccount(ctx, tx, accountID, userID, -amount)
	if err != nil {
		return nil, err
	}

	if err := s.transactionStore.CreateTransaction(ctx, tx, &accountID, nil, amount, idempotencyKey, model.TransactionWithdraw, model.TransactionCompleted); err != nil {
		return nil, err
	}

	response := &dto.BalanceResponse{
		AccountID: account.AccountID,
		Balance:   account.Balance,
		Amount:    amount,
		Message:   "Withdrawal completed successfully",
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	if err := s.idempotencyStore.InsertIdempotency(
		ctx,
		tx,
		userID,
		idempotencyKey,
		http.StatusOK,
		responseBytes,
		expiresAt,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return response, nil
}

func (s *AccountService) Transfer(ctx context.Context, req dto.TransferRequest, userID int) (*dto.TransferResponse, error) {
	if req.Amount <= 0 {
		return nil, errors.New("transfer amount must be greater than zero")
	}
	if req.FromID <= 0 || req.ToID <= 0 {
		return nil, errors.New("invalid account id")
	}
	if req.FromID == req.ToID {
		return nil, errors.New("cannot transfer to same account")
	}

	idempotencyKey, ok := ctx.Value(middleware.IdempotencyKey).(string)
	if !ok || idempotencyKey == "" {
		return nil, errors.New("missing idempotency key")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	existing, err := s.idempotencyStore.GetIdempotency(ctx, tx, userID, idempotencyKey)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if existing != nil {
		var response dto.TransferResponse
		if err := json.Unmarshal(existing.Response, &response); err != nil {
			return nil, err
		}
		return &response, nil
	}

	sender, err := s.accStore.GetAccountByID(ctx, tx, req.FromID)
	if err != nil {
		return nil, fmt.Errorf("sender: %w", err)
	}

	if sender.UserID != userID {
		return nil, errors.New("forbidden")
	}

	if _, err := s.accStore.GetAccountByID(ctx, tx, req.ToID); err != nil {
		return nil, fmt.Errorf("receiver: %w", err)
	}

	if sender.Balance < req.Amount {
		return nil, fmt.Errorf("insufficient balance: have %d need %d", sender.Balance, req.Amount)
	}

	if err := s.accStore.Transfer(ctx, tx, req.FromID, req.ToID, req.Amount); err != nil {
		return nil, err
	}

	if err := s.transactionStore.CreateTransaction(ctx, tx, &req.FromID, &req.ToID, req.Amount, idempotencyKey, model.TransactionTransfer, model.TransactionCompleted); err != nil {
		return nil, err
	}

	response := &dto.TransferResponse{
		FromID:  req.FromID,
		ToID:    req.ToID,
		Amount:  req.Amount,
		Message: "Transfer completed successfully",
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	if err := s.idempotencyStore.InsertIdempotency(
		ctx,
		tx,
		userID,
		idempotencyKey,
		http.StatusOK,
		responseBytes,
		expiresAt,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return response, nil
}

func (s *AccountService) GetAccountByID(ctx context.Context, accountID, userID int) (*dto.AccountResponse, error) {
	if accountID <= 0 {
		return nil, errors.New("invalid account id")
	}

	account, err := s.accStore.GetAccountByID(ctx, s.db, accountID)
	if err != nil {
		return nil, err
	}

	if userID != account.UserID {
		return nil, errors.New("unauthorized")
	}

	return &dto.AccountResponse{
		AccountID: account.AccountID,
		Balance:   account.Balance,
	}, nil
}

func (s *AccountService) DeleteAccount(ctx context.Context, accountID, userID int) (*dto.DeleteAccountResponse, error) {
	if accountID <= 0 {
		return nil, errors.New("invalid account id")
	}

	if err := s.accStore.DeleteAccount(ctx, s.db, accountID, userID); err != nil {
		return nil, err
	}

	return &dto.DeleteAccountResponse{
		Message: "Account deleted successfully",
	}, nil
}
