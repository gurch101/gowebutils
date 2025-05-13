package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/validation"
)

func createRequestWithBody(t *testing.T, method, url string, payload interface{}) *http.Request {
	t.Helper()

	var requestBody []byte

	var err error

	// check if payload is a string
	if strPayload, ok := payload.(string); ok {
		requestBody = []byte(strPayload)
	} else {
		requestBody, err = json.Marshal(payload)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, url, bytes.NewReader(requestBody))
	httputils.SetJSONContentTypeRequestHeader(req)

	return req
}

func CreatePostRequest(t *testing.T, url string, payload interface{}) *http.Request {
	t.Helper()

	return createRequestWithBody(t, http.MethodPost, url, payload)
}

func CreatePatchRequest(t *testing.T, url string, payload interface{}) *http.Request {
	t.Helper()

	return createRequestWithBody(t, http.MethodPatch, url, payload)
}

func CreateGetRequest(t *testing.T, url string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, url, nil)
	httputils.SetJSONContentTypeRequestHeader(req)

	return req
}

func CreateDeleteRequest(url string) *http.Request {
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	httputils.SetJSONContentTypeRequestHeader(req)

	return req
}

func AssertValidationError(t *testing.T, responseRecorder *httptest.ResponseRecorder, expectedErrorField string, expectedErrorMessage string) {
	t.Helper()

	if responseRecorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", responseRecorder.Code)
	}

	var response map[string]interface{}

	err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	errMap, ok := response["errors"].([]interface{})[0].(map[string]interface{})
	if !ok {
		t.Errorf("expected error; got %v", response)
	}

	errorKey, ok := errMap["field"]
	if !ok || errorKey != expectedErrorField {
		t.Errorf("expected error field %s; got %v", expectedErrorField, response)
	}

	errorMessage, ok := errMap["message"]
	if !ok || errorMessage != expectedErrorMessage {
		t.Errorf("expected error message %s; got %v", expectedErrorMessage, response)
	}
}

func NewRouter() *chi.Mux {
	return chi.NewRouter()
}

type ValidationErrorResponse struct {
	Errors []validation.Error `json:"errors"`
}

func StringPtr(s string) *string {
	return &s
}

func BoolPtr(b bool) *bool {
	return &b
}

func IntPtr(i int) *int {
	return &i
}

func Int64Ptr(i int64) *int64 {
	return &i
}
