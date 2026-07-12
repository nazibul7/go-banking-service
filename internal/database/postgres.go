package database

import (
	"database/sql"
	"log"
	"time"
)

func NewPostgresDB(dbURL string) (*sql.DB, error) {
	// Open a database handle (doesn't immediately establish a connection)
	// create a empty connection pool manager
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	// Configuring a connection pool
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping verifies a connection to the database is still alive, establishing a connection if necessary.
	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Connected to DB")
	return db, nil
}
