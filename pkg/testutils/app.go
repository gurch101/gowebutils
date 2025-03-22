package testutils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gurch101/gowebutils/pkg/app"
	"github.com/gurch101/gowebutils/pkg/authutils"
	"github.com/gurch101/gowebutils/pkg/dbutils"
)

type TestApp struct {
	*app.App
	TestRouter *chi.Mux
}

func NewTestApp(t *testing.T) TestApp {
	t.Helper()
	db := SetupTestDB(t)
	mailer := NewMockMailer()
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
