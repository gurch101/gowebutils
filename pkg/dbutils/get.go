package dbutils

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// GetById gets a record from the database by its id.
func GetById(db *sql.DB, tableName string, id int64, fields map[string]any) error {
	if id < 0 {
		return ErrRecordNotFound
	}
	projection := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields))

	for field, dest := range fields {
		projection = append(projection, field)
		args = append(args, dest)
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = $1", strings.Join(projection, ","), tableName)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := db.QueryRowContext(ctx, query, id).Scan(args...)
	if err != nil {
		return WrapDBError(err)
	}
	return nil
}
