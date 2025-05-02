package httputils_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gurch101/gowebutils/pkg/httputils"
)

func TestReadJSON(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			// Create a request with the test body
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()

			// Define a destination struct
			type Dest struct {
				Name string `json:"name"`
			}

			_, err := httputils.ReadJSON[Dest](rr, r)

			// Check the error message
			if tt.expectedError == "" && err != nil {
				t.Errorf("expected no error, got %v", err)
			} else if tt.expectedError != "" && (err == nil || !strings.Contains(err.Error(), tt.expectedError)) {
				t.Errorf("expected error %q, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {
	t.Parallel()

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
			wantBody: `{"name":"Alice","age":30,"email":"alice@example.com"}
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
			wantBody: `{"message":"Resource created"}
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
			name:       "non-marshalable data",
			status:     http.StatusOK,
			data:       make(chan int), // Invalid JSON type
			headers:    nil,
			wantBody:   "null\n",
			wantHeader: map[string]string{"Content-Type": "application/json"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rr := httptest.NewRecorder()

			err := httputils.WriteJSON(rr, tt.status, tt.data, tt.headers)

			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: got %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			gotBody := rr.Body.String()
			if gotBody != tt.wantBody {
				t.Errorf("unexpected body: got %q, want %q", gotBody, tt.wantBody)
			}

			if rr.Code != tt.status {
				t.Errorf("unexpected status: got %d, want %d", rr.Code, tt.status)
			}

			for key, value := range tt.wantHeader {
				if got := rr.Header().Get(key); got != value {
					t.Errorf("unexpected header %q: got %q, want %q", key, got, value)
				}
			}
		})
	}
}
