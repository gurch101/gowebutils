package dbutils

import (
	"testing"
)

func TestInsert(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	t.Run("successful insertion", func(t *testing.T) {
		fields := map[string]any{
			"tenant_name":   "Test Tenant",
			"contact_email": "jane@example.com",
			"plan":          "paid",
			"is_active":     true,
		}

		id, err := Insert(db, "tenants", fields)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if id == nil {
			t.Error("Expected non-nil ID, got nil")
		}
		if *id <= 0 {
			t.Errorf("Expected positive ID, got %d", *id)
		}
	})

	t.Run("empty fields map", func(t *testing.T) {
		fields := map[string]any{}

		id, err := Insert(db, "users", fields)
		if err == nil {
			t.Error("Expected error for empty fields map, got nil")
		}
		if id != nil {
			t.Errorf("Expected nil ID, got %d", *id)
		}
	})

	t.Run("invalid table name", func(t *testing.T) {
		fields := map[string]any{
			"name": "Test User",
		}

		id, err := Insert(db, "nonexistent_table", fields)
		if err == nil {
			t.Error("Expected error for invalid table name, got nil")
		}
		if id != nil {
			t.Errorf("Expected nil ID, got %d", *id)
		}
	})

	t.Run("invalid field name", func(t *testing.T) {
		fields := map[string]any{
			"nonexistent_column": "Test Value",
		}

		id, err := Insert(db, "users", fields)
		if err == nil {
			t.Error("Expected error for invalid field name, got nil")
		}
		if id != nil {
			t.Errorf("Expected nil ID, got %d", *id)
		}
	})
}
