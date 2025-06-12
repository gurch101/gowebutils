package testutils

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetProjectRoot() string {
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
