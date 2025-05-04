package testutils

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	// needed for sqlite3 driver.
	_ "github.com/mattn/go-sqlite3"
)

func seedDB(db *sql.DB) error {
	projectRoot := getProjectRoot()
	dataDir := filepath.Join(projectRoot, "db", "data")

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		return nil
	}

	files, err := os.ReadDir(dataDir)
	if err != nil {
		return fmt.Errorf("failed to read data directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		seedFilePath := filepath.Join(dataDir, file.Name())
		seedFilePath = filepath.Clean(seedFilePath)

		data, err := os.ReadFile(seedFilePath)
		if err != nil {
			return fmt.Errorf("failed to read data file %s: %w", seedFilePath, err)
		}

		_, err = db.Exec(string(data))
		if err != nil {
			return fmt.Errorf("failed to execute data file %s: %w", seedFilePath, err)
		}
	}

	return nil
}

func getProjectRoot() string {
	// Assume the Go module root is the project root, where go.mod is located
	// This will walk up the directory tree to find the go.mod file
	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Failed to get current directory: %v", err))
	}

	// Walk up to find the go.mod file, assuming it's at the root of the project
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		dir = filepath.Dir(dir)
	}
}

func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db := dbutils.Open(":memory:")

	// Apply all migrations
	projectRoot := getProjectRoot()
	migrationDir := filepath.Join(projectRoot, "db", "migrations")

	files, err := os.ReadDir(migrationDir)
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".up.sql") {
			continue
		}

		migrationFilePath := filepath.Join(migrationDir, file.Name())
		filepath.Clean(migrationFilePath)

		migration, err := os.ReadFile(migrationFilePath)
		if err != nil {
			t.Fatalf("Failed to read migration file %s: %v", migrationFilePath, err)
		}

		_, err = db.Exec(string(migration))
		if err != nil {
			t.Fatalf("Failed to execute migration %s: %v", migrationFilePath, err)
		}
	}

	// Seed the database
	err = seedDB(db)
	if err != nil {
		t.Fatalf("Failed to seed database: %v", err)
	}

	return db
}
