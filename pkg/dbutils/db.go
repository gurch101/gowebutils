package dbutils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
)

// Key type to avoid context key collisions.
type ctxKey string

const transactionDepthKey ctxKey = "tx_depth"

var ErrInvalidDBConnectionType = errors.New("invalid DB connection type")

// DB is an interface that both sql.DB and sql.Tx satisfy.
type DB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryRow(query string, args ...interface{}) *sql.Row
}

// WithTransaction manages transactions and supports nesting using savepoints.
func WithTransaction(ctx context.Context, db DB, callback func(tx DB) error) error {
	if dbpool, ok := db.(*DBPool); ok {
		return WithTransaction(ctx, dbpool.WriteDB(), callback)
	}

	depth := getTransactionDepth(ctx)

	// Check if we're already in a transaction
	if tx, ok := db.(*sql.Tx); ok {
		return handleSavepoint(ctx, tx, callback, depth+1)
	}

	// Otherwise, start a new transaction
	tx, err := beginTransaction(ctx, db)
	if err != nil {
		return err
	}

	// Defer a rollback in case of an error
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				slog.ErrorContext(ctx, "db error", "message", fmt.Errorf("failed to rollback transaction: %w", rbErr))
			}
		}
	}()

	ctx = setTransactionDepth(ctx, depth+1)

	err = callback(tx)
	if err != nil {
		return err
	}

	// Commit at the top level
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// beginTransaction starts a new SQL transaction.
func beginTransaction(ctx context.Context, db DB) (*sql.Tx, error) {
	dbConn, ok := db.(*sql.DB)
	if !ok {
		return nil, ErrInvalidDBConnectionType
	}

	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return tx, nil
}

// handleSavepoint creates, rolls back, and releases savepoints for nested transactions.
func handleSavepoint(ctx context.Context, tx *sql.Tx, callback func(tx DB) error, depth int) error {
	savepoint := generateSavepointName(depth)

	if _, err := tx.ExecContext(ctx, "SAVEPOINT "+savepoint); err != nil {
		return fmt.Errorf("failed to create savepoint: %w", err)
	}

	// Increment depth for nested transaction
	ctx = setTransactionDepth(ctx, depth)

	err := callback(tx)
	if err != nil {
		if _, rbErr := tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+savepoint); rbErr != nil {
			slog.ErrorContext(ctx, "db error", "message", fmt.Errorf("failed to rollback savepoint: %w", rbErr))
		}

		return err
	}

	if _, err := tx.ExecContext(ctx, "RELEASE SAVEPOINT "+savepoint); err != nil {
		return fmt.Errorf("failed to release savepoint: %w", err)
	}

	return nil
}

// generateSavepointName generates a savepoint name using depth.
func generateSavepointName(depth int) string {
	return "sp_" + strconv.Itoa(depth)
}

// getTransactionDepth retrieves the current transaction depth from the context.
func getTransactionDepth(ctx context.Context) int {
	if depth, ok := ctx.Value(transactionDepthKey).(int); ok {
		return depth
	}

	return 0
}

// setTransactionDepth updates the transaction depth in the context.
func setTransactionDepth(ctx context.Context, depth int) context.Context {
	return context.WithValue(ctx, transactionDepthKey, depth)
}

func openDB(dsn string) *sql.DB {
	db, err := sql.Open(SqliteDriverName, dsn)
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}

	return db
}

// Open opens a SQLite database file.
func Open(filepath string) *sql.DB {
	return openDB(filepath + "?_foreign_keys=1&_journal=WAL")
}

// OpenWithMode opens a SQLite database file with a specific mode.
func OpenWithMode(filepath string, mode string) *sql.DB {
	return openDB(filepath + "?_foreign_keys=1&_journal=WAL&mode=" + mode)
}
