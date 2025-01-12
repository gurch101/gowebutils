package parser_test

import (
	"errors"
	"testing"

	"github.com/gurch101/gowebutils/pkg/parser"
)

func TestParseJSONMapInt64(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		key     string
		want    int64
		wantErr error
	}{
		{
			name:    "valid int64",
			input:   map[string]any{"count": float64(42)},
			key:     "count",
			want:    42,
			wantErr: nil,
		},
		{
			name:    "invalid type",
			input:   map[string]any{"count": "not a number"},
			key:     "count",
			want:    0,
			wantErr: parser.ErrInvalidMapKey,
		},
		{
			name:    "missing key",
			input:   map[string]any{},
			key:     "count",
			want:    0,
			wantErr: parser.ErrInvalidMapKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseJSONMapInt64(tt.input, tt.key)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseJSONMapInt64() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("ParseJSONMapInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseJSONMapString(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		key     string
		want    string
		wantErr error
	}{
		{
			name:    "valid string",
			input:   map[string]any{"name": "test"},
			key:     "name",
			want:    "test",
			wantErr: nil,
		},
		{
			name:    "invalid type",
			input:   map[string]any{"name": 42},
			key:     "name",
			want:    "",
			wantErr: parser.ErrInvalidMapKey,
		},
		{
			name:    "missing key",
			input:   map[string]any{},
			key:     "name",
			want:    "",
			wantErr: parser.ErrInvalidMapKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseJSONMapString(tt.input, tt.key)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseJSONMapString() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("ParseJSONMapString() = %v, want %v", got, tt.want)
			}
		})
	}
}
