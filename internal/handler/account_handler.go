package handler

import (
	"banking-app/internal/dto"
	"banking-app/internal/middleware"
	"banking-app/internal/service"
	"banking-app/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type AccountHandler struct {
	service *service.AccountService
}

func NewAccountHandler(service *service.AccountService) *AccountHandler {
	return &AccountHandler{service: service}
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.ClaimsKey).(*utils.Claims)

	var req dto.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Balance <= 0 {
		http.Error(w, "balance can not be negative or zero", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	account, err := h.service.CreateAccount(ctx, req.Balance, claims.UserID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "request timed out", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(account); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *AccountHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.ClaimsKey).(*utils.Claims)

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	accounts, err := h.service.GetAccounts(ctx, claims.UserID)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "request timed out", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accounts)
}

func (h *AccountHandler) GetAccountByID(w http.ResponseWriter, r *http.Request) {
	accountID, err := parseID(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	claims := r.Context().Value(middleware.ClaimsKey).(*utils.Claims)

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	account, err := h.service.GetAccountByID(ctx, accountID, claims.UserID)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "request timed out", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}

func (h *AccountHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	accountID, err := parseID(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req dto.BalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Handler-level guard: return 400, not 500, for bad input
	if req.Amount <= 0 {
		http.Error(w, "amount must be greater than zero", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	claims := r.Context().Value(middleware.ClaimsKey).(*utils.Claims)

	response, err := h.service.Deposit(ctx, accountID, claims.UserID, req.Amount)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "request timed out", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *AccountHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	accountID, err := parseID(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req dto.BalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Handler-level guard: return 400, not 500, for bad input
	if req.Amount <= 0 {
		http.Error(w, "amount must be greater than zero", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	claims := r.Context().Value(middleware.ClaimsKey).(*utils.Claims)

	response, err := h.service.Withdraw(ctx, accountID, claims.UserID, req.Amount)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "request timed out", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *AccountHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req dto.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	claims := r.Context().Value(middleware.ClaimsKey).(*utils.Claims)

	response, err := h.service.Transfer(ctx, req, claims.UserID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "request timed out", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *AccountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	accountID, err := parseID(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	claims := r.Context().Value(middleware.ClaimsKey).(*utils.Claims)

	response, err := h.service.DeleteAccount(ctx, accountID, claims.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
