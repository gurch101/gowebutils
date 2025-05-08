package generator

import (
	"fmt"
	"strings"

	"github.com/gurch101/gowebutils/pkg/stringutils"
)

const routesTemplate = `package {{.PackageName}}

import (
	"net/http"

	"github.com/gurch101/gowebutils/pkg/app"
)

func Routes(app *app.App) {
	app.AddProtectedRoute(http.MethodGet, "/{{.KebabCaseTableName}}", NewSearch{{.SingularTitleCaseName}}Controller(app).Search{{.SingularTitleCaseName}}Handler)
	app.AddProtectedRoute(http.MethodPost, "/{{.KebabCaseTableName}}", NewCreate{{.SingularTitleCaseName}}Controller(app).Create{{.SingularTitleCaseName}}Handler)
	app.AddProtectedRoute(http.MethodGet, "/{{.KebabCaseTableName}}/{id}", NewGet{{.SingularTitleCaseName}}ByIDController(app).Get{{.SingularTitleCaseName}}ByIDHandler)
	{{- if .HasUpdate}}
	app.AddProtectedRoute(http.MethodPatch, "/{{.KebabCaseTableName}}/{id}", NewUpdate{{.SingularTitleCaseName}}Controller(app).Update{{.SingularTitleCaseName}}Handler)
	{{- end}}
	app.AddProtectedRoute(http.MethodDelete, "/{{.KebabCaseTableName}}/{id}", NewDelete{{.SingularTitleCaseName}}Controller(app).Delete{{.SingularTitleCaseName}}Handler)
}
`

func RenderRoutesTemplate(moduleName string, schema Table) ([]byte, error) {
	tmpl, err := renderTemplateFile(routesTemplate, map[string]any{
		"PackageName":           schema.Name,
		"KebabCaseTableName":    stringutils.SnakeToKebab(schema.Name),
		"SingularTitleCaseName": stringutils.SnakeToTitle(strings.TrimSuffix(schema.Name, "s")),
		"HasUpdate":             schema.HasUpdateAt(),
	})

	if err != nil {
		return nil, fmt.Errorf("error rendering search template: %w", err)
	}

	return tmpl, nil
}
