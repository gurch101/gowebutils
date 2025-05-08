package generator

import (
	"fmt"
	"strings"

	"github.com/gurch101/gowebutils/pkg/stringutils"
)

const existsTemplate = `package {{.PackageName}}

import (
	"context"

	"github.com/gurch101/gowebutils/pkg/dbutils"
)

func {{.SingularTitleCaseName}}Exists(ctx context.Context, db dbutils.DB, id int64) bool {
	return dbutils.Exists(ctx, db, "{{.Name}}", id)
}
`

func RenderExistsTemplate(moduleName string, schema Table) ([]byte, error) {
	tmpl, err := renderTemplateFile(existsTemplate, map[string]string{
		"PackageName":           schema.Name,
		"Name":                  schema.Name,
		"SingularTitleCaseName": stringutils.SnakeToTitle(strings.TrimSuffix(schema.Name, "s")),
	})
	if err != nil {
		return nil, fmt.Errorf("error rendering exists template: %w", err)
	}

	return tmpl, nil
}
