package stringutils_test

import (
	"testing"

	"github.com/gurch101/gowebutils/pkg/stringutils"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "camel case",
			input:    "helloWorld",
			expected: "hello_world",
		},
		{
			name:     "title case",
			input:    "HelloWorld",
			expected: "hello_world",
		},
		{
			name:     "kebab case",
			input:    "hello-world",
			expected: "hello_world",
		},
		{
			name:     "words",
			input:    "hello world",
			expected: "hello_world",
		},
		{
			name:     "mixed case",
			input:    "Hello World",
			expected: "hello_world",
		},
		{
			name:     "abbreviations",
			input:    "XMLHttpRequest",
			expected: "xml_http_request",
		},
		{
			name:     "id",
			input:    "userID",
			expected: "user_id",
		},
		{
			name:     "extra spaces",
			input:    "  extra  spaces  ",
			expected: "extra_spaces",
		},
		{
			name:     "already snake case",
			input:    "already_snake_case",
			expected: "already_snake_case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringutils.ToSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToSnakeCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
