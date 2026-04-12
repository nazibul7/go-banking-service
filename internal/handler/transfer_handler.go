package handler

import (
	"banking-app/internal/store"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type AccountHandler struct {
	store store.AccountStorer
}

func NewAccountHandler(store store.AccountStorer) *AccountHandler {
	return &AccountHandler{store: store}
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var balance int
	if err := json.NewDecoder(r.Body).Decode(&balance); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if balance < 0 {
		http.Error(w, "balance can not be negative", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	account, err := h.store.CreateAccount(ctx, balance)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
	account, err := h.store.GetAccount(ctx, id)

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
	var depoAmount int
	if err := json.NewDecoder(r.Body).Decode(&depoAmount); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if depoAmount <= 0 {
		http.Error(w, "deposit amount must be greater than zero", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	err = h.store.UpdateAccount(ctx, id, depoAmount)

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
	var withdrawAmount int
	if err := json.NewDecoder(r.Body).Decode(&withdrawAmount); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if withdrawAmount <= 0 {
		http.Error(w, "withdrawal amount must be greater than zero", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	err = h.store.UpdateAccount(ctx, id, -withdrawAmount)

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
	if err := h.store.DeleteAccount(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
