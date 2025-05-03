package dbutils_test

import (
	"context"
	"testing"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/fsutils"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestInsert(t *testing.T) {
	t.Parallel()
	db := testutils.SetupTestDB(t)

	defer fsutils.CloseAndPanic(db)

	fields := map[string]any{
		"tenant_name":   "Test Tenant",
		"contact_email": "jane@example.com",
		"plan":          "paid",
		"is_active":     true,
	}

	id, err := dbutils.Insert(context.Background(), db, "tenants", fields)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if id == nil || *id <= 0 {
		t.Error("Expected non-nil positive ID, got nil")
	}
}

func TestInsert_ErrorHandling(t *testing.T) {
	t.Parallel()
	db := testutils.SetupTestDB(t)

	defer fsutils.CloseAndPanic(db)

	tests := []struct {
		name   string
		table  string
		fields map[string]any
	}{
		{
			name:   "empty fields map",
			table:  "users",
			fields: map[string]any{},
		},
		{
			name:  "invalid table name",
			table: "nonexistent_table",
			fields: map[string]any{
				"name": "Test User",
			},
		},
		{
			name:  "invalid field name",
			table: "users",
			fields: map[string]any{
				"nonexistent_column": "Test Value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := dbutils.Insert(context.Background(), db, tt.table, tt.fields)

			if err == nil {
				t.Error("Expected error, got nil")
			}

			if id != nil {
				t.Errorf("Expected nil ID, got %d", *id)
			}
		})
	}
}
