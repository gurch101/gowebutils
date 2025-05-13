package dbutils

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const getTimeout = 3 * time.Second

// ErrNoGetFilters is returned when no filters are provided to the GetBy function.
var ErrNoGetFilters = errors.New("no filters provided")

// GetByID gets a record from the database by its id.
func GetByID(ctx context.Context, db DB, tableName string, id int64, fields map[string]any) error {
	if id < 0 {
		return ErrRecordNotFound
	}

	return GetBy(ctx, db, tableName, fields, map[string]any{"id": id})
}

// GetBy gets a record from the database with the provided filters.
func GetBy(ctx context.Context, db DB, tableName string, fields map[string]any, filters map[string]any) error {
	if len(filters) == 0 {
		return ErrNoGetFilters
	}

	// Build the SELECT projection
	projection := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields))

	for field, dest := range fields {
		projection = append(projection, field)
		args = append(args, dest)
	}

	// Build the WHERE clause
	whereClauses := make([]string, 0, len(filters))
	whereArgs := make([]any, 0, len(filters))

	for field, value := range filters {
		whereClauses = append(whereClauses, field+" = ?")
		whereArgs = append(whereArgs, value)
	}

	// Construct the full query
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s",
		strings.Join(projection, ","),
		tableName,
		strings.Join(whereClauses, " AND "),
	)

	ctx, cancel := context.WithTimeout(ctx, getTimeout)
	defer cancel()

	err := db.QueryRowContext(ctx, query, whereArgs...).Scan(args...)
	if err != nil {
		return WrapDBError(err)
	}

	return nil
}

func Exists(ctx context.Context, db DB, tableName string, id int64) bool {
	if id < 0 {
		return false
	}

	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1)", tableName)

	ctx, cancel := context.WithTimeout(ctx, getTimeout)
	defer cancel()

	var exists bool

	err := db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false
	}

	return exists
}

// ExistsBy checks if a record exists in the database matching the provided filters.
func ExistsBy(ctx context.Context, db DB, tableName string, filters map[string]any) bool {
	if len(filters) == 0 {
		return false
	}

	// Build the WHERE clause
	whereClauses := make([]string, 0, len(filters))
	whereArgs := make([]any, 0, len(filters))
	argPos := 1

	for field, value := range filters {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", field, argPos))
		whereArgs = append(whereArgs, value)
		argPos++
	}

	// Construct the full query
	query := fmt.Sprintf(
		"SELECT EXISTS(SELECT 1 FROM %s WHERE %s)",
		tableName,
		strings.Join(whereClauses, " AND "),
	)

	ctx, cancel := context.WithTimeout(ctx, getTimeout)
	defer cancel()

	var exists bool

	err := db.QueryRowContext(ctx, query, whereArgs...).Scan(&exists)
	if err != nil {
		return false
	}

	return exists
}
