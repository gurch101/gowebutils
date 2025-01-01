package dbutils

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

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

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Apply all migrations
	projectRoot := getProjectRoot()
	migrationDir := filepath.Join(projectRoot, "db", "migrations")
	dataDir := filepath.Join(projectRoot, "db", "data")

	files, err := os.ReadDir(migrationDir)
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %v", err)
	}

	dataFiles, err := os.ReadDir(dataDir)
	if err != nil {
		t.Fatalf("Failed to read data directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".up.sql") {
			continue
		}

		filePath := filepath.Join(migrationDir, file.Name())
		migration, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read migration file %s: %v", filePath, err)
		}

		_, err = db.Exec(string(migration))
		if err != nil {
			t.Fatalf("Failed to execute migration %s: %v", filePath, err)
		}
	}

	for _, file := range dataFiles {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}
		filePath := filepath.Join(dataDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read data file %s: %v", filePath, err)
		}
		_, err = db.Exec(string(data))
		if err != nil {
			t.Fatalf("Failed to execute data file %s: %v", filePath, err)
		}
	}
	return db
}
