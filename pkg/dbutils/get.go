package dbutils

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const getTimeout = 3 * time.Second

// GetByID gets a record from the database by its id.
func GetByID(ctx context.Context, db DB, tableName string, id int64, fields map[string]any) error {
	if id < 0 {
		return ErrRecordNotFound
	}

	projection := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields))

	for field, dest := range fields {
		projection = append(projection, field)
		args = append(args, dest)
	}

	// #nosec G201
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = $1", strings.Join(projection, ","), tableName)

	ctx, cancel := context.WithTimeout(ctx, getTimeout)
	defer cancel()

	err := db.QueryRowContext(ctx, query, id).Scan(args...)
	if err != nil {
		return WrapDBError(err)
	}

	return nil
}

func Exists(ctx context.Context, db DB, tableName string, id int64) bool {
	if id < 0 {
		return false
	}

	// #nosec G201
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
