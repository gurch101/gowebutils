package dbutils

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
)

// DB is an interface that both sql.DB and sql.Tx satisfy.
type DB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// WithTransaction is a helper function that handles transaction creation, error handling, and rollbacks.
func WithTransaction(ctx context.Context, db *sql.DB, callback func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer a rollback in case of an error
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				slog.ErrorContext(ctx, "db error", "message", fmt.Errorf("failed to rollback transaction: %w", rbErr))
			}
		}
	}()

	err = callback(tx)
	if err != nil {
		return err // Return the error to trigger the rollback
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Open opens a SQLite database file.
func Open(filepath string) (*sql.DB, func()) {
	db, err := sql.Open(SqliteDriverName, filepath+"?_foreign_keys=1&_journal=WAL")
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}

	closer := func() {
		closeErr := db.Close()
		if closeErr != nil {
			panic(closeErr)
		}
	}

	return db, closer
}
