package stringutils_test

import (
	"testing"

	"github.com/gurch101/gowebutils/pkg/stringutils"
)

func TestSnakeToTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic snake case",
			input:    "snake",
			expected: "Snake",
		},
		{
			name:     "multiple words",
			input:    "snake_case",
			expected: "SnakeCase",
		},
		{
			name:     "single uppercase word",
			input:    "UPPER",
			expected: "Upper",
		},
		{
			name:     "starts with underscore",
			input:    "_Response",
			expected: "Response",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "with numbers",
			input:    "user_123_name",
			expected: "User123Name",
		},
		{
			name:     "id field",
			input:    "some_id",
			expected: "SomeID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringutils.SnakeToTitle(tt.input)
			if result != tt.expected {
				t.Errorf("SnakeToTitle(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
