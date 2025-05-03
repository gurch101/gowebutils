package dbutils_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/fsutils"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestDeleteByID(t *testing.T) {
	t.Parallel()
	db := testutils.SetupTestDB(t)

	defer fsutils.CloseAndPanic(db)

	err := dbutils.DeleteByID(context.Background(), db, "users", 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify deletion by attempting to get the record
	var name string
	fields := map[string]any{"user_name": &name}

	err = dbutils.GetByID(context.Background(), db, "users", 1, fields)
	if !errors.Is(err, dbutils.ErrRecordNotFound) {
		t.Errorf("Expected record to be deleted, but got error: %v", err)
	}
}

func TestDelete_ErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		table string
		id    int64
		error error
	}{
		{
			name:  "negative ID",
			table: "users",
			id:    -1,
			error: dbutils.ErrRecordNotFound,
		},
		{
			name:  "non-existent record",
			table: "users",
			id:    999,
			error: dbutils.ErrRecordNotFound,
		},
		{
			name:  "invalid table name",
			table: "nonexistent_table",
			id:    1,
			error: dbutils.ErrNoSuchTable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := testutils.SetupTestDB(t)

			defer fsutils.CloseAndPanic(db)

			err := dbutils.DeleteByID(context.Background(), db, tt.table, tt.id)

			if !errors.Is(err, tt.error) {
				t.Error("Expected error, got nil")
			}
		})
	}
}
