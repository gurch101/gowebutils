package testutils

import (
	"context"
	"embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gurch101/gowebutils/pkg/app"
	"github.com/gurch101/gowebutils/pkg/authutils"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/mailutils"
)

type TestApp struct {
	*app.App
	TestRouter *chi.Mux
}

type Option func(options *options) error

type options struct {
	emailTemplates *embed.FS
}

func WithEmailTemplates(emailTemplates *embed.FS) Option {
	return func(options *options) error {
		options.emailTemplates = emailTemplates

		return nil
	}
}

func NewTestApp(t *testing.T, opts ...Option) TestApp {
	t.Helper()

	var options options
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			t.Fatalf("Failed to apply option: %v", err)
		}
	}

	db := SetupTestDB(t)
	mailer := mailutils.NewMockMailer(mailutils.WithEmailTemplates(options.emailTemplates))
	fileService := NewMockFileService()

	app, err := app.NewApp(
		app.WithDB(db),
		app.WithMailer(mailer),
		app.WithFileService(fileService),
		app.WithGetUserExistsFn(getUserExists),
		app.WithGetOrCreateUserFn(getOrCreateUserFn),
	)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	router := NewRouter()

	return TestApp{
		App:        app,
		TestRouter: router,
	}
}

// MakeAuthenticatedGetRequest issues a GET request to the given path with a default test user as the authenticated user.
// The caller is responsible for ensuring that the test user exists in the database.
func (a *TestApp) MakeAuthenticatedGetRequest(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()

	req := CreateGetRequest(t, path)
	req = authutils.ContextSetUser(req, authutils.User{
		ID:       1,
		TenantID: 1,
		UserName: "doesntmatter",
		Email:    "doesntmatter@example.com",
	})

	return a.MakeRequest(req)
}

// MakeAuthenticatedPostRequest issues a POST request to the given path with a default test user as the authenticated user.
// The caller is responsible for ensuring that the test user exists in the database.
func (a *TestApp) MakeAuthenticatedPostRequest(t *testing.T, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	req := CreatePostRequest(t, path, body)
	req = authutils.ContextSetUser(req, authutils.User{
		ID:       1,
		TenantID: 1,
		UserName: "doesntmatter",
		Email:    "doesntmatter@example.com",
	})

	return a.MakeRequest(req)
}

// MakeAuthenticatedDeleteRequest issues a DELETE request to the given path with a default test user as the authenticated user.
// The caller is responsible for ensuring that the test user exists in the database.
func (a *TestApp) MakeAuthenticatedDeleteRequest(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()

	req := CreateDeleteRequest(path)
	req = authutils.ContextSetUser(req, authutils.User{
		ID:       1,
		TenantID: 1,
		UserName: "doesntmatter",
		Email:    "doesntmatter@example.com",
	})

	return a.MakeRequest(req)
}

// MakeAuthenticatedPatchRequest issues a PATCH request to the given path with a default test user as the authenticated user.
// The caller is responsible for ensuring that the test user exists in the database.
func (a *TestApp) MakeAuthenticatedPatchRequest(t *testing.T, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	req := CreatePatchRequest(t, path, body)
	req = authutils.ContextSetUser(req, authutils.User{
		ID:       1,
		TenantID: 1,
		UserName: "doesntmatter",
		Email:    "doesntmatter@example.com",
	})

	return a.MakeRequest(req)
}

// MakeAuthenticatedPutRequest issues a PUT request to the given path with a default test user as the authenticated user.
// The caller is responsible for ensuring that the test user exists in the database.
func (a *TestApp) MakeAuthenticatedPutRequest(t *testing.T, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	req := CreatePutRequest(t, path, body)
	req = authutils.ContextSetUser(req, authutils.User{
		ID:       1,
		TenantID: 1,
		UserName: "doesntmatter",
		Email:    "doesntmatter@example.com",
	})

	return a.MakeRequest(req)
}

func (a *TestApp) MakeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.TestRouter.ServeHTTP(rr, req)

	return rr
}

func getUserExists(_ context.Context, _ dbutils.DB, _ authutils.User) bool {
	return true
}

func getOrCreateUserFn(_ context.Context, _ dbutils.DB, _ string, _ map[string]any) (authutils.User, error) {
	//nolint:exhaustruct
	return authutils.User{
		ID:    1,
		Email: "test@example.com",
	}, nil
}
