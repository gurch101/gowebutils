package validation

import "testing"

func TestCoalesce(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		ptr        *int
		defaultVal int
		expected   int
	}{
		{"non-nil pointer", new(int), 5, 0},
		{"nil pointer", nil, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Coalesce(tt.ptr, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestValidatorMatches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"valid email", "john.doe@gmail.com", false},
		{"invalid email", "john.doe", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Matches(tt.value, EmailRX, "email", "invalid email")
			if v.HasErrors() != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, v.HasErrors())
			}
		})
	}
}

func TestValidatorCheck(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		condition bool
		expected  bool
	}{
		{"true condition", true, false},
		{"false condition", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Check(tt.condition, "field", "message")
			if v.HasErrors() != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, v.HasErrors())
			}
		})
	}
}

func TestValidatorRequired(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"non-empty value", "value", false},
		{"empty value", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Required(tt.value, "field", "message")
			if v.HasErrors() != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, v.HasErrors())
			}
		})
	}
}
