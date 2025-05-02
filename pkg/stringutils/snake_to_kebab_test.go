package stringutils_test

import (
	"testing"

	"github.com/gurch101/gowebutils/pkg/stringutils"
)

func TestSnakeToKebab(t *testing.T) {
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
			expected: "snake-case",
		},
		{
			name:     "single uppercase word",
			input:    "UPPER",
			expected: "UPPER",
		},
		{
			name:     "starts with underscore",
			input:    "_Response",
			expected: "-Response",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "with numbers",
			input:    "user_123_name",
			expected: "user-123-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringutils.SnakeToKebab(tt.input)
			if result != tt.expected {
				t.Errorf("SnakeToKebab(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
