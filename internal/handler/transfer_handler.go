package handler

import (
	"banking-app/internal/model"
	"banking-app/internal/service"
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
	var req model.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Balance < 0 {
		http.Error(w, "balance can not be negative", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	account, err := h.service.CreateAccount(ctx, req.Balance)
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

func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	account, err := h.service.GetAccount(ctx, id)

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
	id, err := parseID(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req model.AmountRequest
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

	err = h.service.Deposit(ctx, id, req.Amount)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "request timed out", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AccountHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req model.AmountRequest
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

	err = h.service.Withdraw(ctx, id, req.Amount)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "request timed out", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AccountHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req model.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	err := h.service.Transfer(ctx, req)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "request timed out", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AccountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.service.DeleteAccount(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
