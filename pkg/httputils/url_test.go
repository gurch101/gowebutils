package httputils_test

import (
	"testing"

	"github.com/gurch101/gowebutils/pkg/httputils"
)

func TestGetURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		queryParams map[string]string
		expected    string
		expectError bool
	}{
		{
			name:    "valid URL with query params",
			baseURL: "https://example.com",
			queryParams: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expected:    "https://example.com?key1=value1&key2=value2",
			expectError: false,
		},
		{
			name:        "valid URL without query params",
			baseURL:     "https://example.com",
			queryParams: nil,
			expected:    "https://example.com",
			expectError: false,
		},
		{
			name:    "valid URL with empty query params",
			baseURL: "https://example.com",
			queryParams: map[string]string{
				"key1": "",
				"key2": "",
			},
			expected:    "https://example.com?key1=&key2=",
			expectError: false,
		},
		{
			name:        "invalid base URL",
			baseURL:     ":invalid",
			queryParams: map[string]string{"key1": "value1"},
			expected:    "",
			expectError: true,
		},
		{
			name:    "valid URL with special characters in query params",
			baseURL: "https://example.com",
			queryParams: map[string]string{
				"key1": "value 1",
				"key2": "value@2",
			},
			expected:    "https://example.com?key1=value+1&key2=value%402",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := httputils.GetURL(tt.baseURL, tt.queryParams)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}
