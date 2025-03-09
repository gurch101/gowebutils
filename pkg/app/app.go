package app

import (
	"database/sql"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/parser"
)

type Config struct{}

// App is the main application struct.
type App struct {
	DB       *sql.DB
	dbCloser func()
}

// NewApp creates a new instance of the App struct.
func NewApp() *App {
	db, closer := dbutils.Open(parser.ParseEnvStringPanic("DB_FILEPATH"))

	return &App{
		DB:       db,
		dbCloser: closer,
	}
}

// Close closes any resources used by the App.
func (a *App) Close() {
	a.dbCloser()
}
