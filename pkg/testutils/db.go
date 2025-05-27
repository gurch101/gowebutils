package testutils

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/fsutils"

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
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") || !strings.Contains(file.Name(), "test_") {
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

// DumpTable pretty-prints the contents of a database table for debugging purposes.
func DumpTable(t *testing.T, db *dbutils.DBPool, tableName string) {
	t.Helper()

	cols, rowsData, err := getTableData(db, tableName)
	if err != nil {
		t.Logf("Failed to get data for table %s: %v", tableName, err)
		return
	}

	if len(cols) == 0 {
		t.Logf("Table %s is empty or doesn't exist", tableName)
		return
	}

	// Create a tabwriter for pretty output
	var buf bytes.Buffer
	writer := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	//nolint: errcheck
	defer writer.Flush()

	// Header
	_, err = fmt.Fprintf(writer, strings.Join(cols, "\t")+"\n")

	if err != nil {
		t.Errorf("Failed to write header: %v", err)
		return
	}

	// Rows
	for _, row := range rowsData {
		_, err := fmt.Fprintf(writer, strings.Join(row, "\t")+"\n")
		if err != nil {
			t.Errorf("Failed to write row: %v", err)
			return
		}
	}

	err = writer.Flush()
	if err != nil {
		t.Errorf("Failed to flush tabwriter: %v", err)
		return
	}

	// Print result to test log
	t.Logf("\nContents of table %q:\n%s", tableName, buf.String())
}
func getTableData(db *dbutils.DBPool, tableName string) ([]string, [][]string, error) {
	cols, err := getColumnNames(db, tableName)
	if err != nil {
		return nil, nil, err
	}

	rowsData, err := getRowsAsStrings(db, tableName, cols)
	if err != nil {
		return nil, nil, err
	}

	return cols, rowsData, nil
}

func getColumnNames(db *dbutils.DBPool, tableName string) ([]string, error) {
	query := fmt.Sprintf("SELECT * FROM %s LIMIT 0", tableName)
	rows, err := db.Query(query)

	defer fsutils.CloseAndPanic(rows)

	if err != nil {
		return nil, err
	}

	return rows.Columns()
}

func getRowsAsStrings(db *dbutils.DBPool, tableName string, cols []string) ([][]string, error) {
	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	rows, err := db.Query(query)

	defer fsutils.CloseAndPanic(rows)

	if err != nil {
		return nil, err
	}

	var result [][]string

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))

	for i := range cols {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		result = append(result, convertRowToStrings(values))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func convertRowToStrings(values []interface{}) []string {
	row := make([]string, len(values))

	for i, v := range values {
		switch val := v.(type) {
		case nil:
			row[i] = "NULL"
		case []byte:
			row[i] = string(val)
		default:
			row[i] = fmt.Sprintf("%v", val)
		}
	}

	return row
}
