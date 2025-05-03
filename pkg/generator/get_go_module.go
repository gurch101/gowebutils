package generator

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetModuleNameFromGoMod finds and parses the go.mod file to extract the module name.
func GetModuleNameFromGoMod() (string, error) {
	// Find go.mod file in current directory or parent directories
	goModPath, err := findGoModFile(".")
	if err != nil {
		return "", fmt.Errorf("could not find go.mod file: %w", err)
	}

	// Open the go.mod file
	file, err := os.Open(goModPath)
	if err != nil {
		return "", fmt.Errorf("could not open go.mod file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(fmt.Errorf("%w: Error closing file", err))
		}
	}()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			// Extract module name
			moduleName := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			if moduleName == "" {
				return "", errors.New("malformed module line in go.mod")
			}

			return moduleName, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading go.mod file: %w", err)
	}

	return "", errors.New("module declaration not found in go.mod")
}

// findGoModFile searches for go.mod file in the current directory and parent directories.
func findGoModFile(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return goModPath, nil
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// Reached root directory
			break
		}

		dir = parentDir
	}

	return "", errors.New("go.mod file not found")
}
