package dbutils

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// UpdateById updates a record in the database by its id and version.
func UpdateById(db *sql.DB, tableName string, id int64, version int32, fields map[string]any) error {
	if id < 0 || version < 0 {
		return ErrRecordNotFound
	}

	if _, ok := fields["id"]; ok {
		return fmt.Errorf("field 'id' cannot be updated")
	}

	if _, ok := fields["version"]; ok {
		return fmt.Errorf("field 'version' cannot be updated")
	}

	if len(fields) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setClause := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields))

	i := 1
	for field, value := range fields {
		setClause = append(setClause, fmt.Sprintf("%s = $%d", field, i))
		args = append(args, value)
		i++
	}

	setClause = append(setClause, "version = version + 1")
	args = append(args, id, version)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = %d AND version = %d RETURNING version", tableName, strings.Join(setClause, ", "), id, version)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newVersion int32
	err := db.QueryRowContext(ctx, query, args...).Scan(&newVersion)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return ErrEditConflict
		default:
			return WrapDBError(err)
		}
	}
	return nil
}
