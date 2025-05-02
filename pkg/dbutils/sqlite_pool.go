package dbutils

import (
	"context"
	"database/sql"
)

// DBPool is a wrapper around a read/write database connection pool.
type DBPool struct {
	// DB is the read/write database connection pool.
	writeDB *sql.DB
	// readDB is the read-only database connection pool.
	readDB *sql.DB
}

// OpenDBPool opens a new database connection pool.
func OpenDBPool(dsn string) *DBPool {
	writeDB := OpenWithMode(dsn, "rwc")
	writeDB.SetMaxOpenConns(1)

	readDB := OpenWithMode(dsn, "ro")

	return &DBPool{
		writeDB: writeDB,
		readDB:  readDB,
	}
}

func FromDB(db *sql.DB) *DBPool {
	return &DBPool{
		writeDB: db,
		readDB:  db,
	}
}

// Close closes all database connections.
func (d DBPool) Close() {
	if d.writeDB == d.readDB {
		closeErr := d.writeDB.Close()
		if closeErr != nil {
			panic(closeErr)
		}
	} else {
		closeErr := d.writeDB.Close()
		if closeErr != nil {
			panic(closeErr)
		}

		closeErr = d.readDB.Close()
		if closeErr != nil {
			panic(closeErr)
		}
	}
}

// WriteDB returns the write database connection pool.
func (d DBPool) WriteDB() *sql.DB {
	return d.writeDB
}

// ReadDB returns the read-only database connection pool.
func (d DBPool) ReadDB() *sql.DB {
	return d.readDB
}

// WithTransaction executes a callback function within a db transaction.
func (d DBPool) WithTransaction(ctx context.Context, callback func(tx DB) error) error {
	return WithTransaction(ctx, d.writeDB, callback)
}

// Query executes a query with the given arguments.
func (d DBPool) Query(query string, args ...any) (*sql.Rows, error) {
	//nolint: wrapcheck
	return d.readDB.Query(query, args...)
}

// QueryContext executes a query with the given context and arguments.
func (d DBPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	//nolint: wrapcheck
	return d.readDB.QueryContext(ctx, query, args...)
}

// QueryRow executes a query with the given arguments and returns a single row.
func (d DBPool) QueryRow(query string, args ...any) *sql.Row {
	return d.readDB.QueryRow(query, args...)
}

// QueryRowContext executes a query with the given context and arguments and returns a single row.
func (d DBPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.readDB.QueryRowContext(ctx, query, args...)
}

// Exec executes a query with the given arguments.
func (d DBPool) Exec(query string, args ...interface{}) (sql.Result, error) {
	//nolint: wrapcheck
	return d.writeDB.Exec(query, args...)
}

// ExecContext executes a query with the given context and arguments.
func (d DBPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	//nolint: wrapcheck
	return d.writeDB.ExecContext(ctx, query, args...)
}
