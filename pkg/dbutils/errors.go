package dbutils

import (
	"errors"
	"fmt"
	"strings"
)

// ConstraintError represents an error related to database constraints.
type ConstraintErrorType string

// ErrNoSuchTable is returned when a table does not exist.
var ErrNoSuchTable = errors.New("no such table")

// ErrNoSuchColumn is returned when a column does not exist.
var ErrNoSuchColumn = errors.New("no such column")

// ErrNotNullConstraint is returned when a NOT NULL constraint is violated.
var ErrNotNullConstraint = errors.New("not null constraint")

// ErrUniqueConstraint is returned when a UNIQUE constraint is violated.
var ErrUniqueConstraint = errors.New("unique constraint")

// ErrForeignKeyConstraint is returned when a FOREIGN KEY constraint is violated.
var ErrForeignKeyConstraint = errors.New("foreign key constraint")

// ErrCheckConstraint is returned when a CHECK constraint is violated.
var ErrCheckConstraint = errors.New("check constraint")

// ErrRecordNotFound is returned when a query does not return any rows.
var ErrRecordNotFound = errors.New("no rows found")

// ErrEditConflict is returned if there is a data race and a conflicting edit made by another user.
var ErrEditConflict = errors.New("edit conflict")

const (
	notNullPrefix      = "NOT NULL constraint failed: "
	uniquePrefix       = "UNIQUE constraint failed: "
	foreignKeyPrefix   = "FOREIGN KEY constraint failed"
	checkPrefix        = "CHECK constraint failed: "
	noRowsPrefix       = "sql: no rows in result set"
	noSuchTablePrefix  = "no such table: "
	noSuchColumnPrefix = "no such column: "
)

// parseError parses the error message and returns a ConstraintError if the error is related to database constraints.
func parseError(err error) error {
	input := err.Error()

	// Handle specific error types
	switch {
	case strings.HasPrefix(input, notNullPrefix):
		return handleNotNullError(input)
	case strings.HasPrefix(input, uniquePrefix):
		return handleUniqueError(input)
	case strings.HasPrefix(input, foreignKeyPrefix):
		return handleForeignKeyError()
	case strings.HasPrefix(input, checkPrefix):
		return handleCheckError(input)
	case strings.HasPrefix(input, noRowsPrefix):
		return ErrRecordNotFound
	case strings.HasPrefix(input, noSuchTablePrefix):
		return handleNoSuchTableError(input)
	case strings.HasPrefix(input, noSuchColumnPrefix):
		return handleNoSuchColumnError(input)
	default:
		return fmt.Errorf("unhandled error: %w", err)
	}
}

// handleNotNullError handles NOT NULL constraint errors.
func handleNotNullError(input string) error {
	details := strings.TrimPrefix(input, notNullPrefix)

	return fmt.Errorf("%w: %s", ErrNotNullConstraint, details)
}

// handleUniqueError handles UNIQUE constraint errors.
func handleUniqueError(input string) error {
	details := strings.TrimPrefix(input, uniquePrefix)

	return fmt.Errorf("%w: %s", ErrUniqueConstraint, details)
}

// handleForeignKeyError handles FOREIGN KEY constraint errors.
func handleForeignKeyError() error {
	return ErrForeignKeyConstraint
}

// handleCheckError handles CHECK constraint errors.
func handleCheckError(input string) error {
	details := strings.TrimPrefix(input, checkPrefix)

	return fmt.Errorf("%w: %s", ErrCheckConstraint, details)
}

// handleNoSuchTableError handles errors related to tables that do not exist.
func handleNoSuchTableError(input string) error {
	details := strings.TrimPrefix(input, noSuchTablePrefix)

	return fmt.Errorf("%w: %s", ErrNoSuchTable, details)
}

// handleNoSuchColumnError handles errors related to columns that do not exist.
func handleNoSuchColumnError(input string) error {
	details := strings.TrimPrefix(input, noSuchTablePrefix)

	return fmt.Errorf("%w: %s", ErrNoSuchColumn, details)
}

// WrapDBError returns a ConstraintError if the provided error is a database constraint error.
func WrapDBError(err error) error {
	return parseError(err)
}
