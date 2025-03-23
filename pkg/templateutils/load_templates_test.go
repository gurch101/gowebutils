package templateutils_test

import (
	"embed"
	"testing"

	"github.com/gurch101/gowebutils/pkg/templateutils"
)

//go:embed testdata
var testTemplates embed.FS

func TestLoadTemplates(t *testing.T) {
	t.Parallel()

	templates := templateutils.LoadTemplates(testTemplates)
	if len(templates) == 0 {
		t.Errorf("expected templates to be loaded, got none")
	}

	_, ok := templates["test.go.tmpl"]
	if !ok {
		t.Errorf("expected template 'test.go.tmpl' to be loaded, got none")
	}
}
