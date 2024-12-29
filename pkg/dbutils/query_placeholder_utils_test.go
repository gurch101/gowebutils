package dbutils

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestGeneratePlaceholders(t *testing.T) {
	tests := []struct {
		rowNum    int
		numFields int
		expected  string
	}{
		{0, 3, "($1, $2, $3)"},
		{1, 2, "($3, $4)"},
		{2, 4, "($9, $10, $11, $12)"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("RowNum=%d_NumFields=%d", test.rowNum, test.numFields), func(t *testing.T) {
			result := generatePlaceholders(test.rowNum, test.numFields)
			if result != test.expected {
				t.Errorf("expected %q, got %q", test.expected, result)
			}
		})
	}
}

func TestGetChunkSize(t *testing.T) {
	tests := []struct {
		numRows   int
		numFields int
		expected  int
	}{
		{10000, 2, 10000},
		{10000, 10, 6500},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("NumRows=%d_NumFields=%d", test.numRows, test.numFields), func(t *testing.T) {
			result := GetChunkSize(test.numRows, test.numFields)
			if result != test.expected {
				t.Errorf("expected %d, got %d", test.expected, result)
			}
		})
	}
}

func TestProcessInChunks_EmptyArgs(t *testing.T) {
	args := []any{}
	numFields := 2

	callback := func(chunk []any, placeholders string) error {
		t.Fatalf("callback should not be called for empty args")
		return nil
	}

	err := ProcessInChunks(args, 1, numFields, callback)
	if err == nil || err.Error() != "no arguments provided" {
		t.Fatalf("expected error 'no arguments provided', got %v", err)
	}
}
func TestProcessInChunks_InvalidNumFields(t *testing.T) {
	args := []any{1, 2, 3}
	numFields := 0

	callback := func(chunk []any, placeholders string) error {
		t.Fatalf("callback should not be called for invalid numFields")
		return nil
	}

	err := ProcessInChunks(args, 1, numFields, callback)
	if err == nil || err.Error() != "invalid number of fields: 0" {
		t.Fatalf("expected error 'invalid number of fields: 0', got %v", err)
	}
}

func TestProcessInChunks_SingleChunk(t *testing.T) {
	args := []any{1, 2, 3}
	numFields := 2
	var processedChunks [][]any
	var processedPlaceholders []string

	callback := func(chunk []any, placeholders string) error {
		processedChunks = append(processedChunks, chunk)
		processedPlaceholders = append(processedPlaceholders, placeholders)
		return nil
	}

	err := ProcessInChunks(args, 3, numFields, callback)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify results
	if len(processedChunks) != 1 {
		t.Errorf("expected 1 chunk, got %d", len(processedChunks))
	}
	if strings.Join(processedPlaceholders, ",") != "($1, $2),($3, $4),($5, $6)" {
		t.Errorf("expected placeholders ($1, $2),($3, $4),($5, $6), got %s", strings.Join(processedPlaceholders, ","))
	}
}

func TestProcessInChunks_MultipleChunks(t *testing.T) {
	args := []any{1, 2, 3}
	numFields := 2
	chunkSize := 2
	var processedChunks [][]any
	var processedPlaceholders []string

	callback := func(chunk []any, placeholders string) error {
		processedChunks = append(processedChunks, chunk)
		processedPlaceholders = append(processedPlaceholders, placeholders)
		return nil
	}

	err := ProcessInChunks(args, chunkSize, numFields, callback)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(processedChunks) != 2 {
		t.Errorf("expected 2 chunks, got %d", len(processedChunks))
	}
	if len(processedChunks[0]) != 2 || len(processedChunks[1]) != 1 {
		t.Errorf("expected chunks of size 2 and 1, got %v and %v", len(processedChunks[0]), len(processedChunks[1]))
	}
	if strings.Join(processedPlaceholders, ",") != "($1, $2),($3, $4),($5, $6)" {
		t.Errorf("expected placeholders ($1, $2),($3, $4),($5, $6), got %s", strings.Join(processedPlaceholders, ","))
	}
}

func TestProcessInChunks_CallbackError(t *testing.T) {
	args := []any{1, 2, 3}
	numFields := 2
	expectedError := errors.New("callback error")

	callback := func(chunk []any, placeholders string) error {
		return expectedError
	}

	err := ProcessInChunks(args, 2, numFields, callback)
	if !errors.Is(err, expectedError) {
		t.Fatalf("expected error %v, got %v", expectedError, err)
	}
}
