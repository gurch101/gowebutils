package dbutils_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestDeleteByID(t *testing.T) {
	t.Parallel()
	db := testutils.SetupTestDB(t)

	defer func() {
		closeErr := db.Close()
		if closeErr != nil {
			t.Fatalf("Failed to close database connection: %v", closeErr)
		}
	}()

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
	}{
		{
			name:  "negative ID",
			table: "users",
			id:    -1,
		},
		{
			name:  "non-existent record",
			table: "users",
			id:    999,
		},
		{
			name:  "invalid table name",
			table: "nonexistent_table",
			id:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := testutils.SetupTestDB(t)

			defer func() {
				closeErr := db.Close()
				if closeErr != nil {
					t.Fatalf("Failed to close database connection: %v", closeErr)
				}
			}()

			err := dbutils.DeleteByID(context.Background(), db, tt.table, tt.id)

			if err == nil {
				t.Error("Expected error, got nil")
			}
		})
	}
}
