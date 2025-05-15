package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gurch101/gowebutils/pkg/collectionutils"
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

func AssertValidationErrors(t *testing.T, responseRecorder *httptest.ResponseRecorder, expectedErrors validation.ValidationError) {
	t.Helper()

	if responseRecorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", responseRecorder.Code)
	}

	var response validation.ValidationError

	err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response.Errors) != len(expectedErrors.Errors) {
		t.Errorf("Expected %d errors, got %d", len(expectedErrors.Errors), len(response.Errors))
	}

	for _, expectedError := range expectedErrors.Errors {
		ok := collectionutils.Contains(response.Errors, func(error validation.Error) bool {
			return error.Field == expectedError.Field && error.Message == expectedError.Message
		})

		if !ok {
			t.Errorf("Expected error %v", expectedError)
		}
	}
}

func AssertValidationError(t *testing.T, responseRecorder *httptest.ResponseRecorder, expectedErrorField string, expectedErrorMessage string) {
	t.Helper()

	AssertValidationErrors(t, responseRecorder, validation.ValidationError{
		Errors: []validation.Error{
			{
				Field:   expectedErrorField,
				Message: expectedErrorMessage,
			},
		},
	})
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
