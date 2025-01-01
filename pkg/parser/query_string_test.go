package parser

import (
	"net/url"
	"testing"
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
			result := ParseString(tt.qs, tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
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
			result, err := ParseInt(tt.qs, tt.key, tt.defaultValue)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
			if result != tt.expected {
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
		{"default values", url.Values{}, Filters{Page: 1, PageSize: 10, Sort: "id"}},
		{"custom values", url.Values{"page": {"2"}, "pageSize": {"20"}, "sort": {"name"}}, Filters{Page: 2, PageSize: 20, Sort: "name"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f Filters
			err := f.ParseFilters(tt.qs)
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f Filters
			err := f.ParseFilters(tt.qs)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestFilters_Validate(t *testing.T) {
	// This test assumes that the Validate method will be implemented in the future.
	// Currently, it does nothing, so we just call it to ensure it doesn't panic.
	f := Filters{}
	f.Validate()
}
