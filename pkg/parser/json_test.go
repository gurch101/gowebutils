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

func TestWriteJSON(t *testing.T) {
	type testStruct struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	// Test cases
	tests := []struct {
		name       string
		status     int
		data       interface{}
		headers    http.Header
		wantBody   string
		wantHeader map[string]string
		wantErr    bool
	}{
		{
			name:   "valid JSON response with headers",
			status: http.StatusOK,
			data: testStruct{
				Name:  "Alice",
				Age:   30,
				Email: "alice@example.com",
			},
			headers: http.Header{"X-Custom-Header": []string{"CustomValue"}},
			wantBody: `{
	"name": "Alice",
	"age": 30,
	"email": "alice@example.com"
}
`,
			wantHeader: map[string]string{
				"Content-Type":    "application/json",
				"X-Custom-Header": "CustomValue",
			},
			wantErr: false,
		},
		{
			name:   "valid JSON response without additional headers",
			status: http.StatusCreated,
			data: map[string]string{
				"message": "Resource created",
			},
			headers: nil,
			wantBody: `{
	"message": "Resource created"
}
`,
			wantHeader: map[string]string{
				"Content-Type": "application/json",
			},
			wantErr: false,
		},
		{
			name:       "nil data",
			status:     http.StatusOK,
			data:       nil,
			headers:    nil,
			wantBody:   "null\n",
			wantHeader: map[string]string{"Content-Type": "application/json"},
			wantErr:    false,
		},
		{
			name:    "non-marshalable data",
			status:  http.StatusOK,
			data:    make(chan int), // Invalid JSON type
			headers: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock response recorder
			rr := httptest.NewRecorder()

			// Call the function under test
			err := WriteJSON(rr, tt.status, tt.data, tt.headers)

			// Verify the error
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: got %v, wantErr %v", err, tt.wantErr)
			}

			// If an error is expected, skip further checks
			if tt.wantErr {
				return
			}

			// Verify the response body
			gotBody := rr.Body.String()
			if gotBody != tt.wantBody {
				t.Errorf("unexpected body: got %q, want %q", gotBody, tt.wantBody)
			}

			// Verify the response status
			if rr.Code != tt.status {
				t.Errorf("unexpected status: got %d, want %d", rr.Code, tt.status)
			}

			// Verify the headers
			for key, value := range tt.wantHeader {
				if got := rr.Header().Get(key); got != value {
					t.Errorf("unexpected header %q: got %q, want %q", key, got, value)
				}
			}
		})
	}
}
