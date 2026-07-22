package service

import (
	"banking-app/internal/dto"
	"banking-app/internal/store"
	"context"
	"database/sql"
	"errors"
)

type TransactionService struct {
	db               *sql.DB
	accStore         *store.AccountStore
	transactionStore *store.TransactionStore
}

func NewTransactionService(
	db *sql.DB, accStore *store.AccountStore,
	transactionStore *store.TransactionStore,
) *TransactionService {
	return &TransactionService{
		db:               db,
		accStore:         accStore,
		transactionStore: transactionStore,
	}
}

func (s *TransactionService) GetTransactions(
	ctx context.Context,
	userID int,
) ([]dto.TransactionResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	accs, err := s.accStore.GetAccounts(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	for _, acc := range accs {
		if acc.UserID != userID {
			return nil, errors.New("forbidden")
		}
	}

	transactions, err := s.transactionStore.GetTransactions(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.TransactionResponse, 0, len(transactions))

	for _, t := range transactions {
		responses = append(responses, dto.TransactionResponse{
			ID:              t.ID,
			FromAccountID:   t.FromAccountID,
			ToAccountID:     t.ToAccountID,
			Amount:          t.Amount,
			TransactionType: string(t.TransactionType),
			Status:          string(t.Status),
			CreatedAt:       t.CreatedAt,
		})
	}

	return responses, nil
}

func (s *TransactionService) GetTransactionByID(
	ctx context.Context,
	transactionID, userID int,
) (*dto.TransactionResponse, error) {

	if transactionID <= 0 {
		return nil, errors.New("invalid transaction id")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	accs, err := s.accStore.GetAccounts(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	for _, acc := range accs {
		if acc.UserID != userID {
			return nil, errors.New("forbidden")
		}
	}

	t, err := s.transactionStore.GetTransactionByID(ctx, s.db, transactionID, userID)
	if err != nil {
		return nil, err
	}
	return &dto.TransactionResponse{
		ID:              t.ID,
		FromAccountID:   t.FromAccountID,
		ToAccountID:     t.ToAccountID,
		Amount:          t.Amount,
		TransactionType: string(t.TransactionType),
		Status:          string(t.Status),
		CreatedAt:       t.CreatedAt,
	}, nil
}

func (s *TransactionService) GetAccountTransactions(
	ctx context.Context,
	accountID, userID int,
) ([]dto.TransactionResponse, error) {

	if accountID <= 0 {
		return nil, errors.New("invalid account id")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	acc, err := s.accStore.GetAccountByID(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	if acc.UserID != userID {
		return nil, errors.New("forbidden")
	}

	transactions, err := s.transactionStore.GetAccountTransactions(ctx, s.db, accountID, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.TransactionResponse, 0, len(transactions))

	for _, t := range transactions {
		responses = append(responses, dto.TransactionResponse{
			ID:              t.ID,
			FromAccountID:   t.FromAccountID,
			ToAccountID:     t.ToAccountID,
			Amount:          t.Amount,
			TransactionType: string(t.TransactionType),
			Status:          string(t.Status),
			CreatedAt:       t.CreatedAt,
		})
	}

	return responses, nil
}
