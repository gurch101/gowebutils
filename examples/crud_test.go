package main

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestCreateTenant(t *testing.T) {
	t.Parallel()
	app := testutils.NewTestApp(t)
	defer app.Close()

	createTenantController := NewCreateTenantController(app.App)
	app.TestRouter.Post("/tenants", createTenantController.CreateTenantHandler)

	// Define the input JSON for the request
	createTenantRequest := map[string]interface{}{
		"tenantName":   "TestTenant",
		"contactEmail": "test@example.com",
		"plan":         "free",
	}

	req := testutils.CreatePostRequest(t, "/tenants", createTenantRequest)
	rr := app.MakeRequest(req)

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
	err = app.DB().QueryRowContext(context.Background(), "SELECT id FROM tenants WHERE tenant_name = ?", "TestTenant").Scan(&tenantID)
	if err != nil {
		t.Fatalf("Failed to query tenant: %v", err)
	}
}

func TestCreateTenantInvalidPlan(t *testing.T) {
	t.Parallel()
	app := testutils.NewTestApp(t)
	defer app.Close()

	createTenantController := NewCreateTenantController(app.App)
	app.TestRouter.Post("/tenants", createTenantController.CreateTenantHandler)

	// Define the input JSON for the request
	createTenantRequest := map[string]interface{}{
		"tenantName":   "TestTenant",
		"contactEmail": "acme@acme.com",
		"plan":         "invalid",
	}

	req := testutils.CreatePostRequest(t, "/tenants", createTenantRequest)
	rr := app.MakeRequest(req)

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
	t.Parallel()
	app := testutils.NewTestApp(t)
	defer app.Close()

	createTenantController := NewCreateTenantController(app.App)
	app.TestRouter.Post("/tenants", createTenantController.CreateTenantHandler)

	// Define the input JSON for the request
	createTenantRequest := map[string]interface{}{
		"tenantName":   "TestTenant",
		"contactEmail": "acme@acme.com",
		"plan":         "free",
	}

	req := testutils.CreatePostRequest(t, "/tenants", createTenantRequest)
	rr := app.MakeRequest(req)

	// Check the response status code
	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201 Created, got %d", rr.Code)
	}

	req = testutils.CreatePostRequest(t, "/tenants", createTenantRequest)
	rr = app.MakeRequest(req)

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
	t.Parallel()
	app := testutils.NewTestApp(t)
	defer app.Close()

	getTenantController := NewGetTenantController(app.App)
	app.TestRouter.Get("/tenants/{id}", getTenantController.GetTenantHandler)

	req := testutils.CreateGetRequest(t, "/tenants/1")
	rr := app.MakeRequest(req)

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
	t.Parallel()
	app := testutils.NewTestApp(t)
	defer app.Close()

	getTenantController := NewGetTenantController(app.App)
	app.TestRouter.Get("/tenants/{id}", getTenantController.GetTenantHandler)

	req := testutils.CreateGetRequest(t, "/tenants/invalid")
	rr := app.MakeRequest(req)

	// Check the response status code
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 Not Found, got %d", rr.Code)
	}
}

func TestGetTenantHandler_NotFound(t *testing.T) {
	t.Parallel()
	app := testutils.NewTestApp(t)
	defer app.Close()

	getTenantController := NewGetTenantController(app.App)
	app.TestRouter.Get("/tenants/{id}", getTenantController.GetTenantHandler)

	req := testutils.CreateGetRequest(t, "/tenants/9999")
	rr := app.MakeRequest(req)

	// Check the response status code
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 Not Found, got %d", rr.Code)
	}
}

func TestDeleteTenantHandler(t *testing.T) {
	t.Parallel()
	app := testutils.NewTestApp(t)
	defer app.Close()

	deleteTenantController := NewDeleteTenantController(app.App)
	app.TestRouter.Delete("/tenants/{id}", deleteTenantController.DeleteTenantHandler)

	req := testutils.CreateDeleteRequest("/tenants/1")
	rr := app.MakeRequest(req)

	// Check the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Verify that the tenant has been deleted
	var count int
	err := app.DB().QueryRowContext(context.Background(), "SELECT COUNT(*) FROM tenants WHERE id = 1").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query tenant: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected tenant to be deleted, but it still exists")
	}
}

func TestDeleteTenantHandler_InvalidID(t *testing.T) {
	t.Parallel()
	app := testutils.NewTestApp(t)
	defer app.Close()

	deleteTenantController := NewDeleteTenantController(app.App)
	app.TestRouter.Delete("/tenants/{id}", deleteTenantController.DeleteTenantHandler)

	req := testutils.CreateDeleteRequest("/tenants/invalid")
	rr := app.MakeRequest(req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestDeleteTenantHandler_NotFound(t *testing.T) {
	t.Parallel()
	app := testutils.NewTestApp(t)
	defer app.Close()

	deleteTenantController := NewDeleteTenantController(app.App)
	app.TestRouter.Delete("/tenants/{id}", deleteTenantController.DeleteTenantHandler)

	req := testutils.CreateDeleteRequest("/tenants/9999")
	rr := app.MakeRequest(req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestUpdateTenantHandler(t *testing.T) {
	t.Parallel()
	app := testutils.NewTestApp(t)
	defer app.Close()

	updateTenantController := NewUpdateTenantController(app.App)
	app.TestRouter.Patch("/tenants/{id}", updateTenantController.UpdateTenantHandler)

	// Define the input JSON for the update request
	updateTenantRequest := map[string]interface{}{
		"tenantName":   "UpdatedTenant",
		"contactEmail": "updated@example.com",
		"plan":         "paid",
		"isActive":     false,
	}

	// Create a new HTTP request for the update
	req := testutils.CreatePatchRequest(t, "/tenants/1", updateTenantRequest)
	rr := app.MakeRequest(req)

	// Check the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", rr.Code)
	}

	// Check the response body
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["tenantName"] != updateTenantRequest["tenantName"] {
		t.Errorf("Expected tenant name '%s', got '%s'", updateTenantRequest["tenantName"], response["tenantName"])
	}

	// Verify database update
	var updatedTenant struct {
		TenantName   string
		ContactEmail string
		Plan         string
		IsActive     bool
	}
	err = app.DB().QueryRowContext(context.Background(), `SELECT tenant_name, contact_email, plan, is_active FROM tenants WHERE id = 1`).
		Scan(&updatedTenant.TenantName, &updatedTenant.ContactEmail, &updatedTenant.Plan, &updatedTenant.IsActive)
	if err != nil {
		t.Fatalf("Failed to query updated tenant: %v", err)
	}

	if updatedTenant.TenantName != updateTenantRequest["tenantName"] {
		t.Errorf("Expected tenant name '%s', got '%s'", updateTenantRequest["tenantName"], updatedTenant.TenantName)
	}
	if updatedTenant.ContactEmail != updateTenantRequest["contactEmail"] {
		t.Errorf("Expected contact email '%s', got '%s'", updateTenantRequest["contactEmail"], updatedTenant.ContactEmail)
	}
	if updatedTenant.Plan != updateTenantRequest["plan"] {
		t.Errorf("Expected plan '%s', got '%s'", updateTenantRequest["plan"], updatedTenant.Plan)
	}
	if updatedTenant.IsActive != updateTenantRequest["isActive"] {
		t.Errorf("Expected isActive '%v', got '%v'", updateTenantRequest["isActive"], updatedTenant.IsActive)
	}
}

func TestUpdateTenantHandler_InvalidID(t *testing.T) {
	t.Parallel()
	app := testutils.NewTestApp(t)
	defer app.Close()

	updateTenantController := NewUpdateTenantController(app.App)
	app.TestRouter.Patch("/tenants/{id}", updateTenantController.UpdateTenantHandler)

	req := testutils.CreatePatchRequest(t, "/tenants/invalid", map[string]interface{}{})
	rr := app.MakeRequest(req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestUpdateTenantHandler_NotFound(t *testing.T) {
	t.Parallel()
	app := testutils.NewTestApp(t)
	defer app.Close()

	updateTenantController := NewUpdateTenantController(app.App)
	app.TestRouter.Patch("/tenants/{id}", updateTenantController.UpdateTenantHandler)

	req := testutils.CreatePatchRequest(t, "/tenants/9999", map[string]interface{}{})
	rr := app.MakeRequest(req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestUpdateTenantHandler_InvalidRequest(t *testing.T) {
	t.Parallel()
	app := testutils.NewTestApp(t)
	defer app.Close()

	updateTenantController := NewUpdateTenantController(app.App)
	app.TestRouter.Patch("/tenants/{id}", updateTenantController.UpdateTenantHandler)

	req := testutils.CreatePatchRequest(t, "/tenants/1", map[string]interface{}{
		"tenantName":   "UpdatedTenant",
		"contactEmail": "updated@example.com",
		"plan":         "invalid_plan",
		"isActive":     true,
	})
	rr := app.MakeRequest(req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestSearchTenantsHandler(t *testing.T) {
	t.Parallel()
	app := testutils.NewTestApp(t)
	defer app.Close()

	searchTenantsController := NewSearchTenantController(app.App)
	app.TestRouter.Get("/tenants", searchTenantsController.SearchTenantsHandler)

	// Test cases
	testCases := []struct {
		name           string
		queryString    string
		expectedStatus int
		expectedCount  int
		checkResults   func(t *testing.T, response map[string]interface{})
	}{
		{
			name:           "Search all tenants",
			queryString:    "/tenants",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			checkResults:   nil,
		},
		{
			name:           "Search by tenant name",
			queryString:    "/tenants?tenantName=Acme",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResults: func(t *testing.T, response map[string]interface{}) {
				tenants := response["tenants"].([]interface{})
				for _, tenant := range tenants {
					tenantObj := tenant.(map[string]interface{})
					if !strings.Contains(tenantObj["tenantName"].(string), "Acme") {
						t.Errorf("Expected tenant name to contain 'Acme', got %s", tenantObj["tenantName"])
					}
				}
			},
		},
		{
			name:           "Search by plan",
			queryString:    "/tenants?plan=paid",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResults: func(t *testing.T, response map[string]interface{}) {
				tenants := response["tenants"].([]interface{})
				tenant := tenants[0].(map[string]interface{})
				if tenant["plan"] != "paid" {
					t.Errorf("Expected plan to be 'paid', got %s", tenant["plan"])
				}
			},
		},
		{
			name:           "Search by active status",
			queryString:    "/tenants?isActive=false",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			checkResults: func(t *testing.T, response map[string]interface{}) {
				tenants := response["tenants"].([]interface{})
				if len(tenants) != 0 {
					t.Errorf("Expected 0 tenants, got %d", len(tenants))
				}
			},
		},
		{
			name:           "Search by email",
			queryString:    "/tenants?contactEmail=admin@a",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			checkResults: func(t *testing.T, response map[string]interface{}) {
				tenants := response["tenants"].([]interface{})
				for _, tenant := range tenants {
					tenantObj := tenant.(map[string]interface{})
					if !strings.Contains(tenantObj["contactEmail"].(string), "admin") {
						t.Errorf("Expected email to contain 'admin', got %s", tenantObj["contactEmail"])
					}
				}
			},
		},
		{
			name:           "Pagination - page 1, size 2",
			queryString:    "/tenants?page=1&pageSize=2",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			checkResults: func(t *testing.T, response map[string]interface{}) {
				metadata := response["metadata"].(map[string]interface{})
				if metadata["currentPage"].(float64) != 1 {
					t.Errorf("Expected currentPage to be 1, got %v", metadata["currentPage"])
				}
				if metadata["pageSize"].(float64) != 2 {
					t.Errorf("Expected pageSize to be 2, got %v", metadata["pageSize"])
				}
			},
		},
		{
			name:           "Sorting - by tenant name ascending",
			queryString:    "/tenants?sort=tenantName",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			checkResults: func(t *testing.T, response map[string]interface{}) {
				tenants := response["tenants"].([]interface{})
				// Check if sorted
				var names []string
				for _, tenant := range tenants {
					tenantObj := tenant.(map[string]interface{})
					names = append(names, tenantObj["tenantName"].(string))
				}

				sortedNames := make([]string, len(names))
				copy(sortedNames, names)
				sort.Strings(sortedNames)

				for i := range names {
					if names[i] != sortedNames[i] {
						t.Errorf("Expected sorted names, got unsorted")
						break
					}
				}
			},
		},
		{
			name:           "Sorting - by tenant name descending",
			queryString:    "/tenants?sort=-tenantName",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			checkResults: func(t *testing.T, response map[string]interface{}) {
				tenants := response["tenants"].([]interface{})
				// Check if sorted in descending order
				var names []string
				for _, tenant := range tenants {
					tenantObj := tenant.(map[string]interface{})
					names = append(names, tenantObj["tenantName"].(string))
				}

				sortedNames := make([]string, len(names))
				copy(sortedNames, names)
				sort.Sort(sort.Reverse(sort.StringSlice(sortedNames)))

				for i := range names {
					if names[i] != sortedNames[i] {
						t.Errorf("Expected reverse sorted names, got unsorted")
						break
					}
				}
			},
		},
		{
			name:           "Invalid sort field",
			queryString:    "/tenants?sort=invalidField",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
			checkResults:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := testutils.CreateGetRequest(t, tc.queryString)
			rr := app.MakeRequest(req)

			// Check status code
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, rr.Code)
			}

			// For successful requests, check the response content
			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				tenants := response["tenants"].([]interface{})
				if len(tenants) != tc.expectedCount {
					t.Errorf("Expected %d tenants, got %d", tc.expectedCount, len(tenants))
				}

				if tc.checkResults != nil {
					tc.checkResults(t, response)
				}
			}
		})
	}
}
