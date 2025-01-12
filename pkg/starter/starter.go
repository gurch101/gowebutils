// Package starter provides an application server bootstrapper.
package starter

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gurch101/gowebutils/pkg/authutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/parser"
)

const compressionLevel = 5

type Routable interface {
	PublicRoutes(r chi.Router)
	ProtectedRoutes(r chi.Router)
}

type AuthService[T any] interface {
	GetOrCreateUser(ctx context.Context, email string, tokenPayload map[string]any) (T, error)
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
	sessionMiddleware := authutils.GetSessionMiddleware(sessionManager, authService.GetUserExists)
	router := chi.NewRouter()
	router.Use(middleware.RealIP)
	router.Use(middleware.RequestID)
	router.Use(httputils.RateLimitMiddleware)
	router.Use(middleware.RequestLogger(httputils.NewSlogLogFormatter(slog.Default())))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(compressionLevel))
	router.Use(sessionManager.LoadAndSave)

	for _, routable := range routables {
		routable.PublicRoutes(router)
	}

	router.Group(func(r chi.Router) {
		r.Use(sessionMiddleware)

		for _, routable := range routables {
			routable.ProtectedRoutes(r)
		}
	})

	oidcController := authutils.NewOidcController(sessionManager, authService.GetOrCreateUser, oidcConfig)
	oidcController.PublicRoutes(router)
	oidcController.ProtectedRoutes(router)

	fileServer := http.FileServer(http.Dir("./web/static/"))
	router.Handle("/static/*", http.StripPrefix("/static", fileServer))

	err = httputils.ServeHTTP(router, logger)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
