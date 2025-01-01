package dbutils

import (
	"testing"
)

func TestUpdateById(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	t.Run("successful update", func(t *testing.T) {
		fields := map[string]any{
			"user_name": "Jane Doe",
			"email":     "jane@example.com",
		}

		err := UpdateById(db, "users", 1, 1, fields)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify update
		var name, email string
		getFields := map[string]any{
			"user_name": &name,
			"email":     &email,
		}
		err = GetById(db, "users", 1, getFields)
		if err != nil {
			t.Errorf("Failed to verify update: %v", err)
		}
		if name != "Jane Doe" {
			t.Errorf("Expected name 'Jane Doe', got '%s'", name)
		}
		if email != "jane@example.com" {
			t.Errorf("Expected email 'jane@example.com', got '%s'", email)
		}
	})

	t.Run("negative ID", func(t *testing.T) {
		fields := map[string]any{
			"name": "Test User",
		}

		err := UpdateById(db, "users", -1, 1, fields)
		if err != ErrRecordNotFound {
			t.Errorf("Expected ErrRecordNotFound, got %v", err)
		}
	})

	t.Run("negative version", func(t *testing.T) {
		fields := map[string]any{
			"name": "Test User",
		}

		err := UpdateById(db, "users", 1, -1, fields)
		if err != ErrRecordNotFound {
			t.Errorf("Expected ErrRecordNotFound, got %v", err)
		}
	})

	t.Run("attempt to update id field", func(t *testing.T) {
		fields := map[string]any{
			"id": 2,
		}

		err := UpdateById(db, "users", 1, 1, fields)
		if err == nil {
			t.Error("Expected error when updating id field, got nil")
		}
	})

	t.Run("attempt to update version field", func(t *testing.T) {
		fields := map[string]any{
			"version": 2,
		}

		err := UpdateById(db, "users", 1, 1, fields)
		if err == nil {
			t.Error("Expected error when updating version field, got nil")
		}
	})

	t.Run("empty fields map", func(t *testing.T) {
		fields := map[string]any{}

		err := UpdateById(db, "users", 1, 1, fields)
		if err == nil {
			t.Error("Expected error for empty fields map, got nil")
		}
	})

	t.Run("non-existent record", func(t *testing.T) {
		fields := map[string]any{
			"user_name": "Test User",
		}

		err := UpdateById(db, "users", 999, 1, fields)
		if err != ErrEditConflict {
			t.Errorf("Expected ErrEditConflict, got %v", err)
		}
	})

	t.Run("non-existent field", func(t *testing.T) {
		fields := map[string]any{
			"foobar": "Test User",
		}

		err := UpdateById(db, "users", 1, 1, fields)
		if err == nil {
			t.Errorf("Expected error")
		}
	})

	t.Run("version mismatch", func(t *testing.T) {
		fields := map[string]any{
			"user_name": "Test User",
		}

		err := UpdateById(db, "users", 1, 999, fields)
		if err != ErrEditConflict {
			t.Errorf("Expected ErrEditConflict, got %v", err)
		}
	})
}
