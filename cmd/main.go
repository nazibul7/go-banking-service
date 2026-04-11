package main

import (
	"banking-app/internal/app"
	"banking-app/internal/database"
	"banking-app/internal/handler"
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

	store := store.NewAccountStore(db)
	handler := handler.NewAccountHandler(store)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /accounts/create", handler.CreateAccount)
	

	server := app.NewServer(":8080", nil)
	if err := app.RunWithGracefulShutdown(server, 10*time.Second); err != nil {
		log.Fatalf("Server error %v\n", err)
	}
}
