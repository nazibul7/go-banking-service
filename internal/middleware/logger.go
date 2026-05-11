package middleware

import (
	"log"
	"net/http"
	"time"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := r.Context().Value(RequestIDkey)

		next.ServeHTTP(w, r)
		log.Printf("requestID=%s,path=%s, methode=%s,duration=%s", reqID, r.URL.Path, r.Method, time.Since(start))
	})
}
