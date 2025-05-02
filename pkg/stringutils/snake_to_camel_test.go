package stringutils_test

import (
	"testing"

	"github.com/gurch101/gowebutils/pkg/stringutils"
)

func TestSnakeToCamel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic snake case",
			input:    "snake",
			expected: "snake",
		},
		{
			name:     "multiple words",
			input:    "snake_case",
			expected: "snakeCase",
		},
		{
			name:     "single uppercase word",
			input:    "UPPER",
			expected: "upper",
		},
		{
			name:     "starts with underscore",
			input:    "_Response",
			expected: "response",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "with numbers",
			input:    "user_123_name",
			expected: "user123Name",
		},
		{
			name:     "id field",
			input:    "some_id",
			expected: "someId",
		},
		{
			name:     "id field",
			input:    "id",
			expected: "id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringutils.SnakeToCamel(tt.input)
			if result != tt.expected {
				t.Errorf("SnakeToCamel(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
