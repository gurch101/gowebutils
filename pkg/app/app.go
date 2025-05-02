package app

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gurch101/gowebutils/pkg/authutils"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/fsutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/mailutils"
	"github.com/gurch101/gowebutils/pkg/parser"
	"github.com/gurch101/gowebutils/pkg/templateutils"
)

const compressionLevel = 5

var (
	ErrEmailTemplatesNotFound  = errors.New("email templates not found")
	ErrGetUserExistsFnNotFound = errors.New(
		"getUserExistsFn not found. This function is required for session validation")
	ErrGetOrCreateUserFnNotFound = errors.New(
		"getOrCreateUserFn not found. This function is required for user sign in and sign up")
)

// App is the main application struct.
type App struct {
	db                *dbutils.DBPool
	FileService       fsutils.FileService
	Mailer            mailutils.Mailer
	htmlTemplateMap   map[string]*template.Template
	getUserExistsFn   func(ctx context.Context, db dbutils.DB, user authutils.User) bool
	getOrCreateUserFn func(
		ctx context.Context,
		db dbutils.DB,
		email string,
		tokenPayload map[string]any) (authutils.User, error)
	router            *chi.Mux
	sessionMiddleware func(next http.Handler) http.Handler
	sessionManager    *scs.SessionManager
	config            *config
}

type config struct {
	envVars map[string]interface{}
}

func newConfig() *config {
	return &config{envVars: make(map[string]interface{})}
}

func (c *config) getEnvVarString(key string) string {
	val, exists := c.envVars[key]
	if !exists {
		c.envVars[key] = parser.ParseEnvStringPanic(key)
		val = c.envVars[key]
	}

	strVal, ok := val.(string)
	if !ok {
		panic(fmt.Sprintf("expected string value for key %s, but got %T", key, val))
	}

	return strVal
}

func (c *config) getEnvVarInt(key string) int {
	val, exists := c.envVars[key]
	if !exists {
		c.envVars[key] = parser.ParseEnvIntPanic(key)
		val = c.envVars[key]
	}

	intVal, ok := val.(int)
	if !ok {
		panic(fmt.Sprintf("expected int value for key %s, but got %T", key, val))
	}

	return intVal
}

func (c *config) hasEnvVar(key string) bool {
	_, exists := c.envVars[key]
	if exists {
		return true
	}

	if parser.ParseEnvString(key, "") != "" {
		return true
	}

	return false
}

type options struct {
	db                *dbutils.DBPool
	mailer            mailutils.Mailer
	emailTemplateMap  map[string]*template.Template
	htmlTemplateMap   map[string]*template.Template
	fileService       fsutils.FileService
	getUserExistsFn   func(ctx context.Context, db dbutils.DB, user authutils.User) bool
	getOrCreateUserFn func(
		ctx context.Context,
		db dbutils.DB,
		email string,
		tokenPayload map[string]any) (authutils.User, error)
	router *chi.Mux
}

type Option func(options *options) error

func WithDB(db *sql.DB) Option {
	return func(options *options) error {
		options.db = dbutils.FromDB(db)

		return nil
	}
}

func WithMailer(mailer mailutils.Mailer) Option {
	return func(options *options) error {
		options.mailer = mailer

		return nil
	}
}

func WithEmailTemplates(emailTemplates embed.FS) Option {
	return func(options *options) error {
		options.emailTemplateMap = templateutils.LoadTemplates(emailTemplates)

		return nil
	}
}

func WithHTMLTemplates(htmlTemplates embed.FS) Option {
	return func(options *options) error {
		options.htmlTemplateMap = templateutils.LoadTemplates(htmlTemplates)

		return nil
	}
}

func WithGetUserExistsFn(getUserExistsFn func(ctx context.Context, db dbutils.DB, user authutils.User) bool) Option {
	return func(options *options) error {
		options.getUserExistsFn = getUserExistsFn
		if options.getUserExistsFn == nil {
			return ErrGetUserExistsFnNotFound
		}

		return nil
	}
}

func WithGetOrCreateUserFn(getOrCreateUserFn func(
	ctx context.Context,
	db dbutils.DB,
	email string,
	tokenPayload map[string]any) (authutils.User, error),
) Option {
	return func(options *options) error {
		options.getOrCreateUserFn = getOrCreateUserFn
		if options.getOrCreateUserFn == nil {
			return ErrGetOrCreateUserFnNotFound
		}

		return nil
	}
}

func WithFileService(fileService fsutils.FileService) Option {
	return func(options *options) error {
		options.fileService = fileService

		return nil
	}
}

func WithRouter(router *chi.Mux) Option {
	return func(options *options) error {
		options.router = router

		return nil
	}
}

func initDefaultRouter(sessionManager *scs.SessionManager) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RealIP)
	router.Use(middleware.RequestID)
	router.Use(httputils.RateLimitMiddleware)
	router.Use(middleware.RequestLogger(httputils.NewSlogLogFormatter(slog.Default())))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(compressionLevel))
	router.Use(sessionManager.LoadAndSave)

	return router
}

// NewApp creates a new instance of the App struct.
func NewApp(opts ...Option) (*App, error) {
	var options options
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	if options.db == nil {
		db := dbutils.OpenDBPool(parser.ParseEnvStringPanic("DB_FILEPATH"))
		options.db = db
	}

	if options.fileService == nil && parser.ParseEnvString("AWS_S3_REGION", "") != "" {
		fileService := fsutils.NewService(
			parser.ParseEnvStringPanic("AWS_S3_REGION"),
			parser.ParseEnvStringPanic("AWS_S3_BUCKET_NAME"),
			parser.ParseEnvStringPanic("AWS_ACCESS_KEY_ID"),
			parser.ParseEnvStringPanic("AWS_SECRET_ACCESS_KEY"),
		)
		options.fileService = fileService
	}

	if options.mailer == nil && parser.ParseEnvString("SMTP_HOST", "") != "" {
		if options.emailTemplateMap == nil {
			return nil, ErrEmailTemplatesNotFound
		}

		mailer := mailutils.InitMailer(options.emailTemplateMap)
		options.mailer = mailer
	}

	sessionManager := authutils.CreateSessionManager(options.db.WriteDB())
	sessionMiddleware := authutils.GetSessionMiddleware(sessionManager, options.getUserExistsFn, options.db)

	if options.router == nil {
		options.router = initDefaultRouter(sessionManager)
	}

	fileServer := http.FileServer(http.Dir("./web/static/"))
	options.router.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return &App{
		db:                options.db,
		FileService:       options.fileService,
		Mailer:            options.mailer,
		htmlTemplateMap:   options.htmlTemplateMap,
		getUserExistsFn:   options.getUserExistsFn,
		getOrCreateUserFn: options.getOrCreateUserFn,
		router:            options.router,
		sessionMiddleware: sessionMiddleware,
		sessionManager:    sessionManager,
		config:            newConfig(),
	}, nil
}

// AddProtectedRoute adds a route that requires a valid session cookie or jwt to the App.
func (a *App) AddProtectedRoute(method, path string, handler http.HandlerFunc) {
	a.router.With(a.sessionMiddleware, middleware.NoCache).Method(method, path, handler)
}

func (a *App) AddProtectedRouteWithPermission(method, path string, handler http.HandlerFunc, permission string) {
	a.router.With(
		a.sessionMiddleware,
		middleware.NoCache,
		authutils.RequirePermission(permission),
	).Method(method, path, handler)
}

// AddPublicRoute adds a route that does not require a valid session cookie or jwt to the App.
func (a *App) AddPublicRoute(method, path string, handler http.HandlerFunc) {
	a.router.Method(method, path, handler)
}

// GetEnvVarString returns the value of the environment variable with the given key.
// If the environment variable is not set, the app will panic.
func (a *App) GetEnvVarString(key string) string {
	return a.config.getEnvVarString(key)
}

// GetEnvVarInt returns the value of the environment variable with the given key as an integer.
// If the environment variable is not set, the app will panic.
func (a *App) GetEnvVarInt(key string) int {
	return a.config.getEnvVarInt(key)
}

// HasEnvVar returns true if the environment variable with the given key is set.
func (a *App) HasEnvVar(key string) bool {
	return a.config.hasEnvVar(key)
}

// Close closes any resources used by the App.
func (a *App) Close() {
	a.db.Close()
}

// RenderTemplate renders an HTML template with the given name and data.
func (a *App) RenderTemplate(wr io.Writer, name string, data any) error {
	err := a.htmlTemplateMap[name].ExecuteTemplate(wr, name, data)
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}

func (a *App) DB() *dbutils.DBPool {
	return a.db
}

func (a *App) Start() error {
	logger := httputils.InitializeSlog(parser.ParseEnvString("LOG_LEVEL", "info"))

	if parser.ParseEnvString("OIDC_CLIENT_ID", "") != "" {
		oidcController := authutils.CreateOidcController(a.sessionManager,
			func(ctx context.Context, email string, inviteTokenPayload map[string]any) (
				authutils.User, error,
			) {
				return a.getOrCreateUserFn(ctx, a.DB(), email, inviteTokenPayload)
			})
		a.AddPublicRoute("GET", "/login", oidcController.LoginHandler)
		a.AddPublicRoute("GET", "/register", oidcController.RegisterHandler)
		a.AddPublicRoute("GET", "/auth/callback", oidcController.AuthCallback)
		a.AddProtectedRoute("GET", "/logout", oidcController.LogoutHandler)
	}

	err := httputils.ServeHTTP(a.router, logger)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
