package dbutils

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func SeedDb(db *sql.DB) error {
	projectRoot := getProjectRoot()
	dataDir := filepath.Join(projectRoot, "db", "data")

	files, err := os.ReadDir(dataDir)
	if err != nil {
		return fmt.Errorf("failed to read data directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}
		filePath := filepath.Join(dataDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read data file %s: %w", filePath, err)
		}
		_, err = db.Exec(string(data))
		if err != nil {
			return fmt.Errorf("failed to execute data file %s: %w", filePath, err)
		}
	}
	return nil
}
