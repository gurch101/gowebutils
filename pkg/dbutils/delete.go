package dbutils

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// DeleteById deletes a record from the specified table by its ID.
func DeleteById(db *sql.DB, tableName string, id int64) error {
	if id < 0 {
		return ErrRecordNotFound
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
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
