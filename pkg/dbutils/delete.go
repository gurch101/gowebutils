package dbutils

import (
	"context"
	"fmt"
	"time"
)

const deleteTimeout = 3 * time.Second

// DeleteByID deletes a record from the specified table by its ID.
func DeleteByID(ctx context.Context, db DB, tableName string, id int64) error {
	if id < 0 {
		return ErrRecordNotFound
	}

	// #nosec G201
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	result, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return WrapDBError(err)
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
