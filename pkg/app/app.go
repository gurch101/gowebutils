package app

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/fsutils"
	"github.com/gurch101/gowebutils/pkg/mailutils"
	"github.com/gurch101/gowebutils/pkg/parser"
	"github.com/gurch101/gowebutils/pkg/templateutils"
)

var ErrEmailTemplatesNotFound = errors.New("email templates not found")

// App is the main application struct.
type App struct {
	DB              *sql.DB
	FileService     fsutils.FileService
	Mailer          mailutils.Mailer
	htmlTemplateMap map[string]*template.Template
	dbCloser        func()
}

type options struct {
	db               *sql.DB
	dbCloser         func()
	mailer           mailutils.Mailer
	emailTemplateMap map[string]*template.Template
	htmlTemplateMap  map[string]*template.Template
	fileService      fsutils.FileService
}

type Option func(options *options) error

func WithDB(db *sql.DB, closer func()) Option {
	return func(options *options) error {
		options.db = db
		options.dbCloser = closer

		return nil
	}
}

func WithMailer(mailer mailutils.Mailer) Option {
	return func(options *options) error {
		options.mailer = mailer

		return nil
	}
}

func WithEmailTemplates(emailTemplates embed.FS, path string) Option {
	return func(options *options) error {
		options.emailTemplateMap = templateutils.LoadTemplates(emailTemplates, path)

		return nil
	}
}

func WithHTMLTemplates(htmlTemplates embed.FS, path string) Option {
	return func(options *options) error {
		options.htmlTemplateMap = templateutils.LoadTemplates(htmlTemplates, path)

		return nil
	}
}

func WithFileService(fileService fsutils.FileService) Option {
	return func(options *options) error {
		options.fileService = fileService

		return nil
	}
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
		db, closer := dbutils.Open(parser.ParseEnvStringPanic("DB_FILEPATH"))
		options.db = db
		options.dbCloser = closer
	}

	if options.fileService == nil {
		fileService := fsutils.NewService(
			parser.ParseEnvStringPanic("AWS_S3_REGION"),
			parser.ParseEnvStringPanic("AWS_S3_BUCKET_NAME"),
			parser.ParseEnvStringPanic("AWS_ACCESS_KEY_ID"),
			parser.ParseEnvStringPanic("AWS_SECRET_ACCESS_KEY"),
		)
		options.fileService = fileService
	}

	if options.mailer == nil {
		if options.emailTemplateMap == nil {
			return nil, ErrEmailTemplatesNotFound
		}

		mailer := mailutils.InitMailer(options.emailTemplateMap)
		options.mailer = mailer
	}

	return &App{
		DB:              options.db,
		FileService:     options.fileService,
		Mailer:          options.mailer,
		htmlTemplateMap: options.htmlTemplateMap,
		dbCloser:        options.dbCloser,
	}, nil
}

// Close closes any resources used by the App.
func (a *App) Close() {
	a.dbCloser()
}

// RenderTemplate renders an HTML template with the given name and data.
func (a *App) RenderTemplate(wr io.Writer, name string, data any) error {
	err := a.htmlTemplateMap[name].ExecuteTemplate(wr, name, data)
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}
