package middleware

import "net/http"

type contextKey string

const RequestIDkey contextKey = "requestID"

func RequestID() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")

		if id==""{
			id=uuid.New().String()
		}
	})
}
