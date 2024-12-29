package dbutils

import (
	"testing"
)

func TestDeleteById(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	t.Run("successful deletion", func(t *testing.T) {
		err := DeleteById(db, "users", 1)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify deletion by attempting to get the record
		var name string
		fields := map[string]any{"name": &name}
		err = GetById(db, "users", 1, fields)
		if err != ErrRecordNotFound {
			t.Errorf("Expected record to be deleted, but got error: %v", err)
		}
	})

	t.Run("negative ID", func(t *testing.T) {
		err := DeleteById(db, "users", -1)
		if err != ErrRecordNotFound {
			t.Errorf("Expected ErrRecordNotFound, got %v", err)
		}
	})

	t.Run("non-existent record", func(t *testing.T) {
		err := DeleteById(db, "users", 999)
		if err != ErrRecordNotFound {
			t.Errorf("Expected ErrRecordNotFound, got %v", err)
		}
	})

	t.Run("invalid table name", func(t *testing.T) {
		err := DeleteById(db, "nonexistent_table", 1)
		if err == nil {
			t.Error("Expected error for invalid table, got nil")
		}
	})
}
