package testutils

import (
	"testing"

	"github.com/gurch101/gowebutils/pkg/app"
)

func NewTestApp(t *testing.T) *app.App {
	t.Helper()
	db, closer := SetupTestDB(t)
	mailer := NewMockMailer()
	fileService := NewMockFileService()

	app, err := app.NewApp(app.WithDB(db, closer), app.WithMailer(mailer), app.WithFileService(fileService))
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	return app
}
