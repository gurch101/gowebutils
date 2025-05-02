package testutils

import (
	"os"
	"strings"
	"testing"
)

const readWritePermissions = 0o600

func AssertFileEquals(t *testing.T, expectedFilePath, actualFilePath string) {
	t.Helper()

	actual, err := os.ReadFile(actualFilePath)
	if err != nil {
		t.Fatalf("error reading actual file: %v", err)
	}

	expected, err := os.ReadFile(expectedFilePath)
	if err != nil {
		t.Fatalf("error reading expected file: %v", err)
	}

	if strings.TrimSpace(string(expected)) != strings.TrimSpace(string(actual)) {
		t.Fatalf("expected file %s does not match actual file %s", expectedFilePath, actualFilePath)
	}

	os.Remove(actualFilePath)
}

func AssertFileEqualsString(t *testing.T, expectedFilePath, actual string) {
	t.Helper()

	expected, err := os.ReadFile(expectedFilePath)
	if err != nil {
		if err := os.WriteFile(expectedFilePath, []byte(strings.TrimSpace(actual)), readWritePermissions); err != nil {
			t.Fatalf("error writing to expected file: %v", err)
		}

		return
	}

	if strings.TrimSpace(string(expected)) != strings.TrimSpace(actual) {
		t.Fatalf("expected file %s does not match actual contents", expectedFilePath)
	}
}
