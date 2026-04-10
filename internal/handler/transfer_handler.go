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
