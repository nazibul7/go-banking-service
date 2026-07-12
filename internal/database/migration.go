package database

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations applies all pending migration files to the database.
// If a migration has already been applied, it will be skipped.
func RunMigrations(dbURL string) {
	// Create a migration instance. Out of two param it takes-
	// It reads migration files from the migrations folder
	// and connects to the database using dbURL.
	m, err := migrate.New("file://migrations/", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	// Apply all migrations that have not been run yet.
	// Ignore ErrNoChange because it simply means
	// the database is already up to date.
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
	log.Println("migrations applied")
}
