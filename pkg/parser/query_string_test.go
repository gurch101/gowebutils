package parser

import (
	"net/url"
	"testing"

	"gurch101.github.io/go-web/pkg/validation"
)

func TestParseString(t *testing.T) {
	tests := []struct {
		name         string
		qs           url.Values
		key          string
		defaultValue string
		expected     string
	}{
		{"key exists", url.Values{"key": {"value"}}, "key", "default", "value"},
		{"key does not exist", url.Values{}, "key", "default", "default"},
		{"key exists but empty", url.Values{"key": {""}}, "key", "default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseString(tt.qs, tt.key, &tt.defaultValue)
			if *result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, *result)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		name         string
		qs           url.Values
		key          string
		defaultValue int
		expected     int
		expectErr    bool
	}{
		{"key exists", url.Values{"key": {"10"}}, "key", 5, 10, false},
		{"key does not exist", url.Values{}, "key", 5, 5, false},
		{"key exists but empty", url.Values{"key": {""}}, "key", 5, 5, false},
		{"key exists but invalid", url.Values{"key": {"invalid"}}, "key", 5, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseInt(tt.qs, tt.key, &tt.defaultValue)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
			if *result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestFilters_ParseFilters(t *testing.T) {
	tests := []struct {
		name     string
		qs       url.Values
		expected Filters
	}{
		{"default values", url.Values{}, Filters{Page: 1, PageSize: 25, Sort: "id"}},
		{"custom values", url.Values{"page": {"2"}, "pageSize": {"20"}, "sort": {"name"}}, Filters{Page: 2, PageSize: 20, Sort: "name"}},
	}

	v := validation.NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f Filters
			err := f.ParseFilters(tt.qs, v, []string{"id", "name"})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if f.Page != tt.expected.Page || f.PageSize != tt.expected.PageSize || f.Sort != tt.expected.Sort {
				t.Errorf("expected %+v, got %+v", tt.expected, f)
			}
		})
	}
}

func TestFilters_InvalidFilters(t *testing.T) {
	tests := []struct {
		name string
		qs   url.Values
	}{
		{"invalid page", url.Values{"page": {"invalid"}, "pageSize": {"10"}, "sort": {"name"}}},
		{"invalid pageSize", url.Values{"page": {"1"}, "pageSize": {"invalid"}, "sort": {"name"}}},
	}

	v := validation.NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f Filters
			err := f.ParseFilters(tt.qs, v, []string{"id", "name"})
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func TestFilters_InvalidFilterValues(t *testing.T) {
	tests := []struct {
		name string
		qs   url.Values
	}{
		{"invalid page", url.Values{"page": {"-1"}, "pageSize": {"-1"}, "sort": {"invalid"}}},
	}

	v := validation.NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f Filters
			err := f.ParseFilters(tt.qs, v, []string{"id", "name"})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !v.HasErrors() {
				t.Error("expected validation errors, got none")
			}
			if len(v.Errors) != 3 {
				t.Errorf("expected 3 errors, got %d", len(v.Errors))
			}

			var errorFields []string
			for _, err := range v.Errors {
				errorFields = append(errorFields, err.Field)
			}
			if !contains(errorFields, "page") || !contains(errorFields, "page_size") || !contains(errorFields, "sort") {
				t.Errorf("expected errors for page, page_size, and sort, got %v", errorFields)
			}
		})
	}
}
