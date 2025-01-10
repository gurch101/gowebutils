// Package starter provides an application server bootstrapper.
package starter

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gurch101/gowebutils/pkg/authutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/parser"
)

type Routable interface {
	RegisterRoutes(router *httputils.Router)
}

type AuthService[T any] interface {
	GetOrCreateUser(ctx context.Context, email string) (T, error)
	GetUserExists(ctx context.Context, user T) bool
}

func CreateAppServer[T any](authService AuthService[T], db *sql.DB, routables ...Routable) error {
	logger := httputils.InitializeSlog(parser.ParseEnvString("LOG_LEVEL", "info"))

	oidcConfig, err := authutils.CreateOauthConfig(
		parser.ParseEnvStringPanic("OIDC_CLIENT_ID"),
		parser.ParseEnvStringPanic("OIDC_CLIENT_SECRET"),
		parser.ParseEnvStringPanic("OIDC_DISCOVERY_URL"),
		parser.ParseEnvStringPanic("REGISTRATION_URL"),
		parser.ParseEnvStringPanic("LOGOUT_URL"),
		parser.ParseEnvStringPanic("POST_LOGOUT_REDIRECT_URL"),
		parser.ParseEnvStringPanic("OIDC_REDIRECT_URL"),
	)
	if err != nil {
		panic(err)
	}

	sessionManager := authutils.CreateSessionManager(db)

	oidcController := authutils.NewOidcController(sessionManager, authService.GetOrCreateUser, oidcConfig)
	router := httputils.NewRouter(
		authutils.GetSessionMiddleware(sessionManager, authService.GetUserExists),
	)
	router.Use(
		httputils.LoggingMiddleware,
		httputils.RecoveryMiddleware,
		httputils.RateLimitMiddleware,
		httputils.GzipMiddleware,
		sessionManager.LoadAndSave,
	)

	oidcController.RegisterRoutes(router)

	for _, routable := range routables {
		routable.RegisterRoutes(router)
	}

	fileServer := http.FileServer(http.Dir("./web/static/"))
	router.AddStaticRoute("/static/", http.StripPrefix("/static", fileServer))

	err = httputils.ServeHTTP(router.BuildHandler(), logger)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
