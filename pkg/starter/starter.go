// Package starter provides an application server bootstrapper.
package starter

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gurch101/gowebutils/pkg/authutils"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/fsutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/mailutils"
	"github.com/gurch101/gowebutils/pkg/parser"
)

const compressionLevel = 5

// AuthService is an interface that defines the methods that an authentication service must implement.
type AuthService[T any] interface {
	// GetOrCreateUser gets or creates a user based on the email and token payload.
	// It returns the user and an error if one occurred.
	// This is used to sign in or sign up a new user.
	GetOrCreateUser(ctx context.Context, email string, tokenPayload map[string]any) (T, error)
	// GetUserExists checks if a user exists in the database. This is used to validate the user's session.
	GetUserExists(ctx context.Context, user T) bool
}

// AppServer is the main application server struct.
// It contains the database connection, mailer, file service, router, and configuration.
type AppServer[T any] struct {
	DB              *sql.DB
	Mailer          mailutils.Mailer
	FileService     *fsutils.Service
	publicRoutes    []Route
	protectedRoutes []Route
	dbCloser        func()
	config          *config
	htmlTemplateMap map[string]*template.Template
	authService     AuthService[T]
}

// Route is a struct that contains the method, path, and handler function for a route.
type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}

type config struct {
	config map[string]interface{}
}

func newConfig() *config {
	return &config{
		config: make(map[string]interface{}),
	}
}

func (c *config) GetStringPanic(key string) string {
	value, ok := c.config[key]
	if !ok {
		value = parser.ParseEnvStringPanic(key)
		c.config[key] = value
	}
	return value.(string)
}

func (c *config) GetString(key, defaultValue string) string {
	value, ok := c.config[key]
	if !ok {
		value = parser.ParseEnvString(key, defaultValue)
		c.config[key] = value
	}
	return value.(string)
}

func NewAppServer[T any](htmlTemplateMap, emailTemplateMap map[string]*template.Template, authServiceFn func(appserver *AppServer[T]) AuthService[T]) *AppServer[T] {
	config := newConfig()
	db, closer := dbutils.Open(config.GetStringPanic("DB_FILEPATH"))

	mailer := mailutils.InitMailer(emailTemplateMap)

	fileService := fsutils.NewService(
		config.GetStringPanic("AWS_S3_REGION"),
		config.GetStringPanic("AWS_S3_BUCKET_NAME"),
		config.GetStringPanic("AWS_ACCESS_KEY_ID"),
		config.GetStringPanic("AWS_SECRET_ACCESS_KEY"),
	)

	appserver := &AppServer[T]{
		DB:              db,
		Mailer:          mailer,
		FileService:     fileService,
		dbCloser:        closer,
		config:          config,
		htmlTemplateMap: htmlTemplateMap,
	}

	appserver.authService = authServiceFn(appserver)

	return appserver
}

func (s *AppServer[T]) AddPublicRoute(method, path string, handler http.HandlerFunc) {
	s.publicRoutes = append(s.publicRoutes, Route{Method: method, Path: path, Handler: handler})
}

func (s *AppServer[T]) AddProtectedRoute(method, path string, handler http.HandlerFunc) {
	s.protectedRoutes = append(s.protectedRoutes, Route{Method: method, Path: path, Handler: handler})
}

func (s *AppServer[T]) RenderTemplate(w io.Writer, name string, data any) error {
	return s.htmlTemplateMap[name].ExecuteTemplate(w, "index.go.tmpl", nil)
}

func (s *AppServer[T]) GetStringConfig(key string) string {
	return s.config.GetStringPanic(key)
}

func (s *AppServer[T]) Close() {
	s.dbCloser()
}

func (s *AppServer[T]) Start() error {
	defer s.dbCloser()

	logger := httputils.InitializeSlog(s.config.GetString("LOG_LEVEL", "info"))

	sessionManager := authutils.CreateSessionManager(s.DB)
	sessionMiddleware := authutils.GetSessionMiddleware(sessionManager, s.authService.GetUserExists)

	router := chi.NewRouter()
	router.Use(middleware.RealIP)
	router.Use(middleware.RequestID)
	router.Use(httputils.RateLimitMiddleware)
	router.Use(middleware.RequestLogger(httputils.NewSlogLogFormatter(slog.Default())))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(compressionLevel))
	router.Use(sessionManager.LoadAndSave)

	for _, route := range s.publicRoutes {
		router.Method(route.Method, route.Path, route.Handler)
	}

	router.Group(func(r chi.Router) {
		r.Use(middleware.NoCache)
		r.Use(sessionMiddleware)

		for _, route := range s.protectedRoutes {
			r.Method(route.Method, route.Path, route.Handler)
		}
	})

	oidcController := authutils.CreateOidcController(sessionManager, s.authService.GetOrCreateUser)
	oidcController.PublicRoutes(router)
	oidcController.ProtectedRoutes(router)

	fileServer := http.FileServer(http.Dir("./web/static/"))
	router.Handle("/static/*", http.StripPrefix("/static", fileServer))

	err := httputils.ServeHTTP(router, logger)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
