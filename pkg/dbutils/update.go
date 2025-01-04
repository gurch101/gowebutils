package dbutils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrIDNoUpdate = errors.New("field 'id' cannot be updated")

var ErrVersionNoUpdate = errors.New("field 'version' cannot be updated")

var ErrNoFieldsToUpdate = errors.New("no fields to update")

const updateTimeout = 3 * time.Second

// UpdateByID updates a record in the database by its id and version.
func UpdateByID(db *sql.DB, tableName string, id int64, version int32, fields map[string]any) error {
	if id < 0 || version < 0 {
		return ErrRecordNotFound
	}

	if _, ok := fields["id"]; ok {
		return ErrIDNoUpdate
	}

	if _, ok := fields["version"]; ok {
		return ErrVersionNoUpdate
	}

	if len(fields) == 0 {
		return ErrNoFieldsToUpdate
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

	// #nosec G201
	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = %d AND version = %d RETURNING version",
		tableName,
		strings.Join(setClause, ", "),
		id,
		version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), updateTimeout)
	defer cancel()

	var newVersion int32

	err := db.QueryRowContext(ctx, query, args...).Scan(&newVersion)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return WrapDBError(err)
		}
	}

	return nil
}
