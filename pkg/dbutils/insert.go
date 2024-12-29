package dbutils

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Insert inserts a record into the database.
func Insert(db *sql.DB, tableName string, fields map[string]any) (*int64, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("no fields to insert")
	}

	columns := make([]string, 0, len(fields))
	values := make([]any, 0, len(fields))
	placeholders := make([]string, 0, len(fields))
	i := 1
	for field, value := range fields {
		columns = append(columns, field)
		values = append(values, value)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		i++
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id", tableName, strings.Join(columns, ","), strings.Join(placeholders, ","))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var id int64
	err := db.QueryRowContext(ctx, query, values...).Scan(&id)
	if err != nil {
		return nil, WrapDBError(err)
	}
	return &id, nil
}
