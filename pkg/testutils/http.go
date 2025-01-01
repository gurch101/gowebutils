package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func CreatePostRequest(t *testing.T, url string, payload interface{}) *http.Request {
	requestBody, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func AssertError(t *testing.T, resp map[string]interface{}, expectedErrorField string, expectedErrorMessage string) {
	error, ok := resp["errors"].([]interface{})[0].(map[string]interface{})
	if !ok {
		t.Errorf("expected error; got %v", resp)
	}
	errorKey, ok := error["field"]
	if !ok || errorKey != expectedErrorField {
		t.Errorf("expected error field %s; got %v", expectedErrorField, resp)
	}
	errorMessage, ok := error["message"]
	if !ok || errorMessage != expectedErrorMessage {
		t.Errorf("expected error message %s; got %v", expectedErrorMessage, resp)
	}
}
