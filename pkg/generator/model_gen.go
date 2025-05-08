package generator

import (
	"fmt"
	"strings"

	"github.com/gurch101/gowebutils/pkg/stringutils"
)

const modelTemplate = `package {{.PackageName}}

import (
	{{if .IncludeTime}}"time"{{end}}
	{{if .IncludeValidation}}"github.com/gurch101/gowebutils/pkg/validation"{{end}}
)

{{- range .UniqueFields}}
var Err{{.TitleCaseName}}AlreadyExists = validation.Error{
	Field:   "{{.JSONName}}",
	Message: "{{.HumanName}} already exists",
}
{{- end}}
{{- range .ForeignKeys}}
var Err{{.SingularTitleCaseTableName}}NotFound = validation.Error{
	Field:   "{{.JSONName}}",
	Message: "{{.HumanTableName}} not found",
}
{{- end}}

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
	uniqueFields := []UniqueField{}
	includeTime := false
	includeValidation := len(schema.ForeignKeys) > 0

	for _, field := range schema.Fields {
		sanitizedName := getSanitizedName(field.Name)

		if field.DataType == SQLDatetime || field.DataType == SQLTimestamp {
			includeTime = true
		}

		if IsRequestField(field) {
			if hasUniqueConstraint(field.Constraints) {
				uniqueFields = append(uniqueFields, newUniqueField(field))
				includeValidation = true
			}

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
		IncludeValidation:     includeValidation,
		TitleCaseTableName:    stringutils.SnakeToTitle(schema.Name),
		SingularTitleCaseName: stringutils.SnakeToTitle(strings.TrimSuffix(schema.Name, "s")),
		SingularCamelCaseName: stringutils.SnakeToCamel(strings.TrimSuffix(schema.Name, "s")),
		ModelFields:           modelFields,
		Fields:                fields,
		UniqueFields:          uniqueFields,
		ForeignKeys:           schema.ForeignKeys,
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
