package main

import (
	"banking-app/internal/app"
	"banking-app/internal/database"
	"banking-app/internal/handler"
	"banking-app/internal/middleware"
	"banking-app/internal/service"
	"banking-app/internal/store"
	"net/http"

	"log"
	"time"
)

func main() {
	connStr := `postgres://myuser:mypassword@localhost:5433/mydb?sslmode=disable`
	database.RunMigrations(connStr)

	db, err := database.NewPostgresDB(connStr)
	if err != nil {
		log.Fatalf("DB connection error: %v\n", err)
	}
	// Ensure the connection is closed when main exits
	defer db.Close()

	accStore := store.NewAccountStore(db)
	accService := service.NewAccountService(accStore)
	accHandler := handler.NewAccountHandler(accService)

	authStores := store.NewAuthStore(db)
	refreshStore := store.NewRefreshTokenStore(db)
	tsStore := store.NewTxStore(db)
	authService := service.NewAuthService(authStores, refreshStore, tsStore)
	authHandler := handler.NewAuthHandler(authService)

	mux := http.NewServeMux()

	muxHandler := middleware.Logger(mux)
	muxHandler = middleware.RequestID(muxHandler)
	muxHandler = middleware.Recoverer(muxHandler)

	mux.HandleFunc("POST /signup", authHandler.Signup)
	mux.HandleFunc("POST /signin", authHandler.Signin)
	mux.HandleFunc("POST /refresh", authHandler.Refresh)
	mux.Handle("POST /logout", middleware.Auth(http.HandlerFunc(authHandler.Logout)))

	mux.Handle("POST /account", middleware.Auth(http.HandlerFunc(accHandler.CreateAccount)))
	mux.Handle("GET /account/{id}", middleware.Auth(http.HandlerFunc(accHandler.GetAccount)))
	mux.Handle("PATCH /account/{id}/deposit", middleware.Auth(http.HandlerFunc(accHandler.Deposit)))
	mux.Handle("PATCH /account/{id}/withdraw", middleware.Auth(http.HandlerFunc(accHandler.Withdraw)))
	mux.Handle("DELETE /account/{id}", middleware.Auth(http.HandlerFunc(accHandler.DeleteAccount)))
	mux.Handle("POST /account/transfer", middleware.Auth(http.HandlerFunc(accHandler.Transfer)))

	server := app.NewServer(":8080", muxHandler)
	if err := app.RunWithGracefulShutdown(server, 10*time.Second); err != nil {
		log.Fatalf("Server error %v\n", err)
	}
}
