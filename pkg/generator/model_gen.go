package generator

import (
	"fmt"
	"strings"

	"github.com/gurch101/gowebutils/pkg/stringutils"
)

const modelTemplate = `package {{.PackageName}}

import (
	{{if .IncludeTime}}"time"{{end}}
)

type {{.SingularCamelCaseName}}Model struct {
	{{- range .ModelFields}}
	{{.TitleCaseName}} {{.GoType}}
	{{- end}}
}

func newCreate{{.SingularTitleCaseName}}Model(
	{{- range .Fields}}
	{{.JSONName}} {{.GoType}},
	{{- end}}
) *{{.SingularCamelCaseName}}Model {
	return &{{.SingularCamelCaseName}}Model{
		{{- range .Fields}}
		{{.TitleCaseName}}: {{.JSONName}},
		{{- end}}
	}
}

func (m *{{.SingularCamelCaseName}}Model) Field(field string) interface{} {
	switch field {
	{{- range .ModelFields}}
	case "{{.Name}}":
		return &m.{{.TitleCaseName}}
	{{- end}}
	default:
		return nil
	}
}
`

func newModelTemplateData(moduleName string, schema Table) modelTemplateData {
	modelFields := []ModelField{}
	fields := []RequestField{}
	includeTime := false

	for _, field := range schema.Fields {
		sanitizedName := getSanitizedName(field.Name)

		if field.DataType == SQLDatetime || field.DataType == SQLTimestamp {
			includeTime = true
		}

		if field.Name != "id" && field.Name != "version" && field.Name != "created_at" && field.Name != "updated_at" {
			fields = append(fields, RequestField{
				Name:          field.Name,
				TitleCaseName: stringutils.SnakeToTitle(sanitizedName),
				JSONName:      stringutils.SnakeToCamel(sanitizedName),
				HumanName:     stringutils.SnakeToHuman(sanitizedName),
				GoType:        field.DataType.GoType(),
			})
		}

		modelFields = append(modelFields, newModelField(sanitizedName, field))
	}

	return modelTemplateData{
		PackageName:           schema.Name,
		Name:                  schema.Name,
		ModuleName:            moduleName,
		IncludeTime:           includeTime,
		TitleCaseTableName:    stringutils.SnakeToTitle(schema.Name),
		SingularTitleCaseName: stringutils.SnakeToTitle(strings.TrimSuffix(schema.Name, "s")),
		SingularCamelCaseName: strings.ToLower(stringutils.SnakeToCamel(strings.TrimSuffix(schema.Name, "s"))),
		ModelFields:           modelFields,
		Fields:                fields,
	}
}

func RenderModelTemplate(moduleName string, schema Table) ([]byte, error) {
	data := newModelTemplateData(moduleName, schema)

	createTemplate, err := renderTemplateFile(modelTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("error rendering model template: %w", err)
	}

	return createTemplate, nil
}
