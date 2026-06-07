package middleware

import (
	"context"
	"net/http"
)

const IdempotencyKey string = "idempotency_key"

func Idempotency(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			http.Error(w, "idempotency key missing", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), IdempotencyKey, key)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
