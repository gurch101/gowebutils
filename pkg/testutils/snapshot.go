package testutils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const readWritePermissions = 0o600

func AssertFileEquals(t *testing.T, expectedFilePath, actualFilePath string) {
	t.Helper()

	actualFilePath = filepath.Clean(actualFilePath)

	actual, err := os.ReadFile(actualFilePath)
	if err != nil {
		t.Fatalf("error reading actual file: %v", err)
	}

	expectedFilePath = filepath.Clean(expectedFilePath)

	expected, err := os.ReadFile(expectedFilePath)
	if err != nil {
		t.Fatalf("error reading expected file: %v", err)
	}

	if strings.TrimSpace(string(expected)) != strings.TrimSpace(string(actual)) {
		t.Fatalf("expected file %s does not match actual file %s", expectedFilePath, actualFilePath)
	}

	err = os.Remove(actualFilePath)
	if err != nil {
		t.Fatalf("error removing actual file: %v", err)
	}
}

func AssertFileEqualsString(t *testing.T, expectedFilePath, actual string) {
	t.Helper()

	expectedFilePath = filepath.Clean(expectedFilePath)
	expected, err := os.ReadFile(expectedFilePath)

	if err != nil {
		if err := os.WriteFile(expectedFilePath, []byte(strings.TrimSpace(actual)), readWritePermissions); err != nil {
			t.Fatalf("error writing to expected file: %v", err)
		}

		return
	}

	if strings.TrimSpace(string(expected)) != strings.TrimSpace(actual) {
		if err := os.WriteFile(expectedFilePath+".new", []byte(strings.TrimSpace(actual)), readWritePermissions); err != nil {
			t.Fatalf("error writing to expected file: %v", err)
		}

		t.Fatalf("expected file %s does not match actual contents", expectedFilePath)
	}
}
