package repository

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// RunMigrations applies schema changes from migrations folder
func RunMigrations(dsn string) error {
	// goose requires standard *sql.DB
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// Setup dialect
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	// Run migrations from directory
	return goose.Up(db, "migrations")
}