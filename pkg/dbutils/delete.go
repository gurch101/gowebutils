package dbutils

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const deleteTimeout = 3 * time.Second

// DeleteByID deletes a record from the specified table by its ID.
func DeleteByID(db *sql.DB, tableName string, id int64) error {
	if id < 0 {
		return ErrRecordNotFound
	}

	// #nosec G201
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)

	ctx, cancel := context.WithTimeout(context.Background(), deleteTimeout)
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
