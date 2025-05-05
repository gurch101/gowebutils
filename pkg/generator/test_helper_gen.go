package generator

import (
	"fmt"
	"strings"

	"github.com/gurch101/gowebutils/pkg/stringutils"
)

const testHelperTemplate = `package {{.PackageName}}

import (
	"context"
	"testing"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func CreateTest{{.SingularTitleCaseName}}Request(t *testing.T) Create{{.SingularTitleCaseName}}Request {
	t.Helper()

	return Create{{.SingularTitleCaseName}}Request{
		{{- range .Fields}}
		{{- if or (eq .GoType "int") (eq .GoType "int64")}}
		{{.TitleCaseName}}: 1,
		{{- else if eq .GoType "bool"}}
		{{.TitleCaseName}}: true,
		{{- else if .IsEmail}}
		{{.TitleCaseName}}: "test@example.com",
		{{- else}}
		{{.TitleCaseName}}: "{{.TitleCaseName}}",
		{{- end}}
		{{- end}}
	}
}

func CreateTest{{.SingularTitleCaseName}}(t *testing.T, db dbutils.DB) (int64, Create{{.SingularTitleCaseName}}Request) {
	t.Helper()

	createReq := CreateTest{{.SingularTitleCaseName}}Request(t)

	ID, err := Create{{.SingularTitleCaseName}}(context.Background(), db, &createReq)

	if err != nil {
		t.Fatal(err)
	}

	return *ID, createReq
}

func CreateTestUpdate{{.SingularTitleCaseName}}Request(t *testing.T) Update{{.SingularTitleCaseName}}Request {
	t.Helper()

	return Update{{.SingularTitleCaseName}}Request{
	{{- range .Fields}}
	{{- if eq .GoType "int64"}}
	{{.TitleCaseName}}: testutils.Int64Ptr(2),
	{{- else if eq .GoType "int"}}
	{{.TitleCaseName}}: testutils.IntPtr(2),
	{{- else if eq .GoType "bool"}}
	{{.TitleCaseName}}: testutils.BoolPtr(false),
	{{- else if .IsEmail}}
	{{.TitleCaseName}}: testutils.StringPtr("newtest@example.com"),
	{{- else}}
	{{.TitleCaseName}}: testutils.StringPtr("new{{.TitleCaseName}}"),
	{{- end}}
	{{- end}}
	}
}
`

func newTestHelperTemplateData(moduleName string, schema Table) testHelperTemplateData {
	fields := []RequestField{}

	for _, field := range schema.Fields {
		sanitizedName := getSanitizedName(field.Name)

		if IsRequestField(field) {
			fields = append(fields, RequestField{
				Name:          field.Name,
				TitleCaseName: stringutils.SnakeToTitle(sanitizedName),
				JSONName:      stringutils.SnakeToCamel(sanitizedName),
				HumanName:     stringutils.SnakeToHuman(sanitizedName),
				GoType:        field.DataType.GoType(),
				IsEmail:       isEmail(sanitizedName),
			})
		}
	}

	return testHelperTemplateData{
		PackageName:           schema.Name,
		Name:                  schema.Name,
		ModuleName:            moduleName,
		TitleCaseTableName:    stringutils.SnakeToTitle(schema.Name),
		SingularTitleCaseName: stringutils.SnakeToTitle(strings.TrimSuffix(schema.Name, "s")),
		Fields:                fields,
	}
}

func RenderTestHelperTemplate(moduleName string, schema Table) ([]byte, error) {
	data := newTestHelperTemplateData(moduleName, schema)

	createTemplate, err := renderTemplateFile(testHelperTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("error rendering test helper template: %w", err)
	}

	return createTemplate, nil
}
