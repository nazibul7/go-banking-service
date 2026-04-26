package main

import (
	"banking-app/internal/app"
	"banking-app/internal/database"
	"banking-app/internal/handler"
	"banking-app/internal/middleware"
	"banking-app/internal/store"
	"net/http"

	"log"
	"time"
)

func main() {
	connStr := `postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable`
	database.RunMigrations(connStr)

	db, err := database.NewPostgresDB(connStr)
	if err != nil {
		log.Fatalf("DB connection error: %v\n", err)
	}
	// Ensure the connection is closed when main exits
	defer db.Close()

	store := store.NewAccountStore(db)
	handler := handler.NewAccountHandler(store)

	mux := http.NewServeMux()

	muxHandler := middleware.Logger(mux)
	muxHandler = middleware.RequestID(muxHandler)
	muxHandler = middleware.Recoverer(muxHandler)

	mux.HandleFunc("POST /account/create", handler.CreateAccount)
	mux.HandleFunc("GET /account/{id}", handler.GetAccount)
	mux.HandleFunc("PATCH /account/{id}/deposit", handler.Deposit)
	mux.HandleFunc("PATCH /account/{id}/withdraw", handler.Withdraw)
	mux.HandleFunc("DELETE /account/{id}", handler.DeleteAccount)

	server := app.NewServer(":8080", mux)
	if err := app.RunWithGracefulShutdown(server, 10*time.Second); err != nil {
		log.Fatalf("Server error %v\n", err)
	}
}
