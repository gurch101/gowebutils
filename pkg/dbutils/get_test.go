package dbutils

import (
	"testing"
)

func TestGetById(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	t.Run("successful retrieval", func(t *testing.T) {
		var name, email string
		fields := map[string]any{
			"name":  &name,
			"email": &email,
		}

		err := GetById(db, "users", 1, fields)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if name != "John Doe" {
			t.Errorf("Expected name 'John Doe', got '%s'", name)
		}
		if email != "john@example.com" {
			t.Errorf("Expected email 'john@example.com', got '%s'", email)
		}
	})

	t.Run("negative ID", func(t *testing.T) {
		var name string
		fields := map[string]any{
			"name": &name,
		}

		err := GetById(db, "users", -1, fields)
		if err != ErrRecordNotFound {
			t.Errorf("Expected ErrRecordNotFound, got %v", err)
		}
	})

	t.Run("non-existent record", func(t *testing.T) {
		var name string
		fields := map[string]any{
			"name": &name,
		}

		err := GetById(db, "users", 999, fields)
		if err != ErrRecordNotFound {
			t.Errorf("Expected ErrRecordNotFound, got %v", err)
		}
	})

	t.Run("empty fields map", func(t *testing.T) {
		fields := map[string]any{}

		err := GetById(db, "users", 1, fields)
		if err == nil {
			t.Error("Expected error for empty fields map, got nil")
		}
	})
}
