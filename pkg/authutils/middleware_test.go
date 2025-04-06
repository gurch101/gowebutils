package authutils_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gurch101/gowebutils/pkg/authutils"
)

func TestRequirePermission(t *testing.T) {
	tests := []struct {
		name               string
		userPermissions    map[string]bool
		requiredPermission string
		isAdmin            bool
		expectedStatus     int
	}{
		{
			name:               "User has required permission",
			userPermissions:    map[string]bool{"read": true},
			requiredPermission: "read",
			isAdmin:            false,
			expectedStatus:     http.StatusOK,
		},
		{
			name:               "User does not have required permission",
			userPermissions:    map[string]bool{"write": true},
			requiredPermission: "read",
			isAdmin:            false,
			expectedStatus:     http.StatusForbidden,
		},
		{
			name:               "User is admin",
			userPermissions:    map[string]bool{"doesntmatter": true},
			requiredPermission: "doesntmatter",
			isAdmin:            true,
			expectedStatus:     http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock user and context
			user := authutils.User{
				ID:          1,
				TenantID:    1,
				UserName:    "test",
				Email:       "test123@test.com",
				IsAdmin:     tt.isAdmin,
				Permissions: tt.userPermissions,
			}

			handlerCallback := authutils.RequirePermission(tt.requiredPermission)
			handler := handlerCallback(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			// Create a request and response recorder
			req := httptest.NewRequest(http.MethodGet, "/api/foobar", nil)
			req = authutils.ContextSetUser(req, user)
			rr := httptest.NewRecorder()

			// Call the middleware
			handler.ServeHTTP(rr, req)

			// Check the response status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
