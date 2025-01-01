package parser

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadJSON(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		expectedError string
	}{
		{"valid JSON", `{"name":"John"}`, ""},
		{"invalid JSON syntax", `{"name":`, "body contains badly-formed JSON"},
		{"unexpected EOF", `{"name":"John"`, "body contains badly-formed JSON"},
		{"incorrect JSON type", `{"name":123}`, "body contains incorrect JSON type for field \"name\""},
		{"empty body", ``, "body must not be empty"},
		{"unknown field", `{"unknown":"field"}`, "body contains unknown key \"unknown\""},
		{"body too large", `{"name":"` + strings.Repeat("a", 1_048_577) + `"}`, "body must not be larger than 1048576 bytes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request with the test body
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()

			// Define a destination struct
			var dst struct {
				Name string `json:"name"`
			}

			// Call ReadJSON
			err := ReadJSON(w, r, &dst)

			// Check the error message
			if tt.expectedError == "" && err != nil {
				t.Errorf("expected no error, got %v", err)
			} else if tt.expectedError != "" && (err == nil || !strings.Contains(err.Error(), tt.expectedError)) {
				t.Errorf("expected error %q, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestReadJSON_InvalidUnmarshalError(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"John"}`))
	w := httptest.NewRecorder()

	// Pass a non-pointer to ReadJSON to trigger json.InvalidUnmarshalError
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic, got none")
		}
	}()
	ReadJSON(w, r, struct{}{})
}
