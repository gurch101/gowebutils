package dbutils_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestUpdateByID(t *testing.T) {
	t.Parallel()
	db := testutils.SetupTestDB(t)

	defer db.Close()

	fields := map[string]any{
		"user_name": "Jane Doe",
		"email":     "jane@example.com",
	}

	err := dbutils.UpdateByID(context.Background(), db, "users", 1, 1, fields)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify update
	var name, email string
	getFields := map[string]any{
		"user_name": &name,
		"email":     &email,
	}

	err = dbutils.GetByID(context.Background(), db, "users", 1, getFields)
	if err != nil {
		t.Errorf("Failed to verify update: %v", err)
	}

	if name != "Jane Doe" {
		t.Errorf("Expected name 'Jane Doe', got '%s'", name)
	}

	if email != "jane@example.com" {
		t.Errorf("Expected email 'jane@example.com', got '%s'", email)
	}
}

func TestUpdateByID_ErrorHandling(t *testing.T) {
	t.Parallel()
	db := testutils.SetupTestDB(t)

	defer db.Close()

	tests := []struct {
		name     string
		table    string
		id       int64
		version  int32
		fields   map[string]any
		expected error
	}{
		{
			name:     "negative ID",
			table:    "users",
			id:       -1,
			version:  1,
			fields:   map[string]any{"name": "Test User"},
			expected: dbutils.ErrRecordNotFound,
		},
		{
			name:     "negative version",
			table:    "users",
			id:       1,
			version:  -1,
			fields:   map[string]any{"name": "Test User"},
			expected: dbutils.ErrRecordNotFound,
		},
		{
			name:     "attempt to update id field",
			table:    "users",
			id:       1,
			version:  1,
			fields:   map[string]any{"id": 2},
			expected: dbutils.ErrIDNoUpdate,
		},
		{
			name:     "attempt to update version field",
			table:    "users",
			id:       1,
			version:  1,
			fields:   map[string]any{"version": 2},
			expected: dbutils.ErrVersionNoUpdate,
		},
		{
			name:     "empty fields map",
			table:    "users",
			id:       1,
			version:  1,
			fields:   map[string]any{},
			expected: dbutils.ErrNoFieldsToUpdate,
		},
		{
			name:     "non-existent field",
			table:    "users",
			id:       1,
			version:  1,
			fields:   map[string]any{"foobar": "Test User"},
			expected: dbutils.ErrNoSuchColumn,
		},
		{
			name:     "non-existent record",
			table:    "users",
			id:       999,
			version:  1,
			fields:   map[string]any{"user_name": "Test User"},
			expected: dbutils.ErrEditConflict,
		},
		{
			name:     "version mismatch",
			table:    "users",
			id:       1,
			version:  999,
			fields:   map[string]any{"user_name": "Test User"},
			expected: dbutils.ErrEditConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dbutils.UpdateByID(context.Background(), db, tt.table, tt.id, tt.version, tt.fields)
			if !errors.Is(err, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, err)
			}
		})
	}
}
