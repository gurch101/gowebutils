package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gurch101.github.io/go-web/pkg/dbutils"
	"gurch101.github.io/go-web/pkg/testutils"
)

func TestCreateTenant(t *testing.T) {
	db := dbutils.SetupTestDB(t)

	// Create the TenantController instance with the test database
	tenantController := NewTenantController(db)

	// Define the input JSON for the request
	createTenantRequest := map[string]interface{}{
		"tenantName":   "TestTenant",
		"contactEmail": "test@example.com",
		"plan":         "free",
	}

	// Create a new HTTP request
	req := testutils.CreatePostRequest(t, "/tenants", createTenantRequest)

	// Record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request through the controller
	tenantController.GetMux().ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201 Created, got %d", rr.Code)
	}

	// Check the response body
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["id"] == nil {
		t.Error("Expected non-nil ID, got nil")
	}

	var tenantID int64
	err = db.QueryRow("SELECT id FROM tenants WHERE tenant_name = ?", "TestTenant").Scan(&tenantID)
	if err != nil {
		t.Fatalf("Failed to query tenant: %v", err)
	}
}

func TestCreateTenantInvalidPlan(t *testing.T) {
	db := dbutils.SetupTestDB(t)

	// Create the TenantController instance with the test database
	tenantController := NewTenantController(db)

	// Define the input JSON for the request
	createTenantRequest := map[string]interface{}{
		"tenantName":   "TestTenant",
		"contactEmail": "acme@acme.com",
		"plan":         "invalid",
	}

	// Create a new HTTP request
	req := testutils.CreatePostRequest(t, "/tenants", createTenantRequest)

	// Record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request through the controller
	tenantController.GetMux().ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 Created, got %d", rr.Code)
	}

	// Check the response body
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	testutils.AssertError(t, response, "plan", "Invalid plan")
}

func TestCreateTenant_DuplicateTenant(t *testing.T) {
	db := dbutils.SetupTestDB(t)

	// Create the TenantController instance with the test database
	tenantController := NewTenantController(db)

	// Define the input JSON for the request
	createTenantRequest := map[string]interface{}{
		"tenantName":   "TestTenant",
		"contactEmail": "acme@acme.com",
		"plan":         "free",
	}

	// Create a new HTTP request
	req := testutils.CreatePostRequest(t, "/tenants", createTenantRequest)

	// Record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request through the controller
	tenantController.GetMux().ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201 Created, got %d", rr.Code)
	}

	req = testutils.CreatePostRequest(t, "/tenants", createTenantRequest)
	rr = httptest.NewRecorder()
	tenantController.GetMux().ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request, got %d", rr.Code)
	}

	// Check the response body
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	testutils.AssertError(t, response, "tenantName", "This tenant is already registered")
}

func TestGetTenantHandler(t *testing.T) {
	db := dbutils.SetupTestDB(t)

	// Create the TenantController instance with the test database
	tenantController := NewTenantController(db)

	// Create a new HTTP request
	req := testutils.CreateGetRequest(t, "/tenants/1")

	// Record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request through the controller
	tenantController.GetMux().ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", rr.Code)
	}

	// Check the response body
	var response GetTenantResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ID != 1 {
		t.Errorf("Expected tenant ID 1, got %d", response.ID)
	}
	if response.TenantName != "Acme" {
		t.Errorf("Expected tenant name 'Acme', got '%s'", response.TenantName)
	}
	if response.ContactEmail != "admin@acme.com" {
		t.Errorf("Expected contact email 'admin@acme.com', got '%s'", response.ContactEmail)
	}
	if response.Plan != Free {
		t.Errorf("Expected plan 'free', got '%s'", response.Plan)
	}
	if !response.IsActive {
		t.Errorf("Expected tenant to be active")
	}
}

func TestGetTenantHandler_InvalidID(t *testing.T) {
	db := dbutils.SetupTestDB(t)

	// Create the TenantController instance with the test database
	tenantController := NewTenantController(db)

	// Create a new HTTP request with an invalid ID
	req := httptest.NewRequest(http.MethodGet, "/tenants/invalid", nil)

	// Record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request through the controller
	tenantController.GetMux().ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 Not Found, got %d", rr.Code)
	}
}

func TestGetTenantHandler_NotFound(t *testing.T) {
	db := dbutils.SetupTestDB(t)

	// Create the TenantController instance with the test database
	tenantController := NewTenantController(db)

	// Create a new HTTP request for a non-existent tenant
	req := httptest.NewRequest(http.MethodGet, "/tenants/9999", nil)

	// Record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request through the controller
	tenantController.GetMux().ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 Not Found, got %d", rr.Code)
	}
}
