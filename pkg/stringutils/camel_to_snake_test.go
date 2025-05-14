//nolint:dupl
package stringutils_test

import (
	"testing"

	"github.com/gurch101/gowebutils/pkg/stringutils"
)

func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic camel case",
			input:    "camelCase",
			expected: "camel_case",
		},
		{
			name:     "multiple words",
			input:    "ThisIsATest",
			expected: "this_is_a_test",
		},
		{
			name:     "single lowercase word",
			input:    "lower",
			expected: "lower",
		},
		{
			name:     "single uppercase word",
			input:    "UPPER",
			expected: "u_p_p_e_r",
		},
		{
			name:     "consecutive uppercase letters",
			input:    "APIResponse",
			expected: "a_p_i_response",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "with numbers",
			input:    "User123Name",
			expected: "user123_name",
		},
		{
			name:     "with id",
			input:    "UserID",
			expected: "user_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringutils.CamelToSnake(tt.input)
			if result != tt.expected {
				t.Errorf("CamelToSnake(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
