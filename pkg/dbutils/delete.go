package dbutils

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const deleteTimeout = 3 * time.Second

// ErrNoDeleteFilters is returned when no filters are provided to the DeleteBy function.
var ErrNoDeleteFilters = errors.New("no filters provided")

// DeleteByID deletes a record from the specified table by its ID.
func DeleteByID(ctx context.Context, db DB, tableName string, id int64) error {
	if id < 0 {
		return ErrRecordNotFound
	}

	rowsAffected, err := DeleteBy(ctx, db, tableName, map[string]any{"id": id})

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// DeleteBy deletes records from the specified table matching the provided filters.
func DeleteBy(ctx context.Context, db DB, tableName string, filters map[string]any) (int, error) {
	if len(filters) == 0 {
		return 0, ErrNoDeleteFilters
	}

	// Build the WHERE clause
	whereClauses := make([]string, 0, len(filters))
	whereArgs := make([]any, 0, len(filters))

	for field, value := range filters {
		whereClauses = append(whereClauses, field+" = ?")
		whereArgs = append(whereArgs, value)
	}

	// Construct the full query
	// #nosec G201 - tableName is not user input in normal usage
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s",
		tableName,
		strings.Join(whereClauses, " AND "),
	)

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	result, err := db.ExecContext(ctx, query, whereArgs...)
	if err != nil {
		return 0, WrapDBError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, WrapDBError(err)
	}

	if rowsAffected == 0 {
		return 0, nil
	}

	return int(rowsAffected), nil
}
