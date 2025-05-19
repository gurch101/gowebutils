package generator

import (
	"fmt"
	"strings"

	"github.com/gurch101/gowebutils/pkg/stringutils"
)

const searchHandlerTemplate = `package {{.PackageName}}

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gurch101/gowebutils/pkg/app"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/parser"
	"github.com/gurch101/gowebutils/pkg/validation"
)

type Search{{.SingularTitleCaseName}}Controller struct {
	app *app.App
}

func NewSearch{{.SingularTitleCaseName}}Controller(app *app.App) *Search{{.SingularTitleCaseName}}Controller {
	return &Search{{.SingularTitleCaseName}}Controller{app: app}
}

type Search{{.SingularTitleCaseName}}Request struct {
	{{- range .Fields}}
	{{.TitleCaseName}} *{{.GoType}}
	{{- end}}
	parser.Filters
}

type Search{{.SingularTitleCaseName}}Response struct {
	Metadata parser.PaginationMetadata ` + "`" + `json:"metadata"` + "`" + `
	Data     []Search{{.SingularTitleCaseName}}ResponseData ` + "`" + `json:"data"` + "`" + `
}

type Search{{.SingularTitleCaseName}}ResponseData struct {
	{{- range .ModelFields}}
	{{.TitleCaseName}} {{.GoType}} ` + "`" + `json:"{{.CamelCaseName}}"` + "`" + `
	{{- end}}
}

// List{{.SingularTitleCaseName}} godoc
//
//	@Summary		List {{.HumanName}}
//	@Description	get {{.HumanName}}
//	@Tags			{{.HumanName}}
//	@Accept			json
//	@Produce		json
{{- range .Fields}}
//	@Param 			{{.JSONName}} query {{.GoType}} false "{{.JSONName}}"
{{- end}}
//	@Param			fields query string false "csv list of fields to include. By default all fields are included"
//	@Param      page query int false "page number" minimum(1) default(1)
//	@Param			pageSize	query		int		false	"page size" minimum(1)  maximum(100) default(25)
//	@Param			sort	query		string	false	"sort by field. e.g. field1,-field2"
//	@Success		200	{object}		Search{{.SingularTitleCaseName}}Response
//	@Failure		400,500	{object}	httputils.ErrorResponse
//	@Router			/{{.KebabCaseTableName}} [get]
func (tc *Search{{.SingularTitleCaseName}}Controller) Search{{.SingularTitleCaseName}}Handler(w http.ResponseWriter, r *http.Request) {
	queryString := r.URL.Query()
	request := &Search{{.SingularTitleCaseName}}Request{
		{{- range .Fields}}
		{{.TitleCaseName}}: parser.ParseQS{{if eq .GoType "bool"}}Bool{{else if (eq .GoType "int64")}}Int64{{else if (eq .GoType "int")}}Int{{else}}String{{end}}(queryString, "{{.JSONName}}", nil),
		{{- end}}
	}

	v := validation.NewValidator()
	request.ParseQSMetadata(queryString, v, []string{
		{{- range .ModelFields}}
		"{{.CamelCaseName}}",
		{{- end}}
	}, []string{
		"id",
		"-id",
		{{- range .Fields}}
		"{{.JSONName}}",
		"-{{.JSONName}}",
		{{- end}}
	},)

	if v.HasErrors() {
		httputils.FailedValidationResponse(w, r, v.Errors)
		return
	}

	response, err := Search{{.TitleCaseTableName}}(r.Context(), tc.app.DB(), request)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)
		return
	}

	filteredResponse, err := parser.StructsToFilteredMaps(response.Data, request.Fields)

	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
		return
	}

	err = httputils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"metadata": response.Metadata,
		"data":     filteredResponse,
	}, nil)

	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}


func Search{{.TitleCaseTableName}}(
	ctx context.Context,
	db dbutils.DB,
	search{{.SingularTitleCaseName}}Request *Search{{.SingularTitleCaseName}}Request,
) (*Search{{.SingularTitleCaseName}}Response, error) {
	models, pagination, err := find{{.TitleCaseTableName}}(ctx, db, search{{.SingularTitleCaseName}}Request)
	if err != nil {
		return nil, err
	}

	return &Search{{.SingularTitleCaseName}}Response{
		Metadata: pagination,
		Data:     models,
	}, nil
}

func find{{.TitleCaseTableName}}(
	ctx context.Context,
	db dbutils.DB,
	request *Search{{.SingularTitleCaseName}}Request) ([]Search{{.SingularTitleCaseName}}ResponseData, parser.PaginationMetadata, error) {
	var models []Search{{.SingularTitleCaseName}}ResponseData
	var totalRecords int

	dbFields := dbutils.BuildSearchSelectFields("{{.Name}}", request.Fields, nil)

	err := dbutils.NewQueryBuilder(db).
		Select(
			dbFields...
		).
		From("{{.Name}}").
		{{range $i, $field := .Fields}}
			{{if eq $i 0}}
			Where("{{$.Name}}.{{$field.Name}} = ?", request.{{$field.TitleCaseName}}).
			{{else}}
			AndWhere("{{$.Name}}.{{$field.Name}} = ?", request.{{$field.TitleCaseName}}).
			{{end}}
		{{end}}
		OrderBy("{{.Name}}."+request.Sort).
		Page(request.Page, request.PageSize).
		QueryContext(ctx, func(rows *sql.Rows) error {
			model, numRecords, err := Scan{{.SingularTitleCaseName}}Record(rows, dbFields)

			if err != nil {
				return err
			}

			models = append(models, model)
			totalRecords = numRecords

			return nil
		})

	if err != nil {
		return nil, parser.PaginationMetadata{}, dbutils.WrapDBError(err)
	}

	metadata := parser.ParsePaginationMetadata(totalRecords, request.Page, request.PageSize)
	return models, metadata, nil
}

func Scan{{.SingularTitleCaseName}}Record(rows *sql.Rows, dbFields []string) (Search{{.SingularTitleCaseName}}ResponseData, int, error) {
	var m Search{{.SingularTitleCaseName}}ResponseData
	var totalRecords int

	fieldsToBindTo := make([]interface{}, len(dbFields))
	fieldsToBindTo[0] = &totalRecords

	for i, field := range dbFields[1:] {
		switch field {
			{{- range .ModelFields}}
			case "{{$.Name}}.{{.Name}}":
				fieldsToBindTo[i+1] = &m.{{.TitleCaseName}}
			{{- end}}
		}
	}

	err := rows.Scan(
		fieldsToBindTo...,
	)

	if err != nil {
		return m, 0, err
	}

	return m, totalRecords, nil
}
`

const searchHandlerTestTemplate = `package {{.PackageName}}_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"{{.ModuleName}}/internal/{{.PackageName}}"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestSearch{{.SingularTitleCaseName}}(t *testing.T) {
    t.Parallel()

    t.Run("successful search", func(t *testing.T) {
        app := testutils.NewTestApp(t)
        defer app.Close()

				ID, _ := {{.PackageName}}.CreateTest{{.SingularTitleCaseName}}(t, app.DB())

        controller := {{.PackageName}}.NewSearch{{.SingularTitleCaseName}}Controller(app.App)
        app.TestRouter.Get("/{{.KebabCaseTableName}}", controller.Search{{.SingularTitleCaseName}}Handler)

        req := testutils.CreateGetRequest(t, "/{{.KebabCaseTableName}}")

        rr := app.MakeRequest(req)

        if rr.Code != http.StatusOK {
            t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
        }

        var response struct {
            Data []{{.PackageName}}.Search{{.SingularTitleCaseName}}ResponseData ` + "`" + `json:"data"` + "`" + `
        }
        err := json.Unmarshal(rr.Body.Bytes(), &response)
        if err != nil {
            t.Fatal(err)
        }

        if len(response.Data) != 1 {
            t.Fatalf("expected 1 {{.HumanName}}, got %d", len(response.Data))
        }

				actualRecord, err := {{.PackageName}}.Get{{.SingularTitleCaseName}}ByID(context.Background(), app.DB(), ID)
				if err != nil {
					t.Fatal(err)
				}

        {{- range .ModelFields}}
        if response.Data[0].{{.TitleCaseName}} != actualRecord.{{.TitleCaseName}} {
            t.Errorf("expected {{.TitleCaseName}} to be %v, got %v", actualRecord.{{.TitleCaseName}}, response.Data[0].{{.TitleCaseName}})
        }
        {{- end}}
    })

    t.Run("bad sort parameter", func(t *testing.T) {
        app := testutils.NewTestApp(t)
        defer app.Close()

        controller := {{.PackageName}}.NewSearch{{.SingularTitleCaseName}}Controller(app.App)
        app.TestRouter.Get("/{{.KebabCaseTableName}}", controller.Search{{.SingularTitleCaseName}}Handler)

        req := testutils.CreateGetRequest(t, "/{{.KebabCaseTableName}}?sort=invalid")
        rr := app.MakeRequest(req)

				testutils.AssertValidationError(t, rr, "sort", "invalid sort value")
    })

    t.Run("bad field parameter", func(t *testing.T) {
        app := testutils.NewTestApp(t)
        defer app.Close()

        controller := {{.PackageName}}.NewSearch{{.SingularTitleCaseName}}Controller(app.App)
        app.TestRouter.Get("/{{.KebabCaseTableName}}", controller.Search{{.SingularTitleCaseName}}Handler)

        req := testutils.CreateGetRequest(t, "/{{.KebabCaseTableName}}?fields=invalidField")
        rr := app.MakeRequest(req)

				testutils.AssertValidationError(t, rr, "fields", "invalid field: invalidField")
    })

		t.Run("single field", func(t *testing.T) {
			app := testutils.NewTestApp(t)
			defer app.Close()

			{{.PackageName}}.CreateTest{{.SingularTitleCaseName}}(t, app.DB())

			controller := {{.PackageName}}.NewSearch{{.SingularTitleCaseName}}Controller(app.App)
			app.TestRouter.Get("/{{.KebabCaseTableName}}", controller.Search{{.SingularTitleCaseName}}Handler)

			req := testutils.CreateGetRequest(t, "/{{.KebabCaseTableName}}?fields=id")
			rr := app.MakeRequest(req)

			if rr.Code != http.StatusOK {
					t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
			}

			if !strings.Contains(rr.Body.String(), "id") {
					t.Errorf("expected response to contain {{.PackageName}} id, got %s", rr.Body.String())
			}
    })
}
`

func newSearchHandlerTemplateData(moduleName string, schema Table) searchHandlerTemplateData {
	fields := []RequestField{}
	modelFields := []ModelField{}

	for _, field := range schema.Fields {
		if IsRequestField(field) {
			fields = append(fields, RequestField{
				Name:          field.Name,
				TitleCaseName: stringutils.SnakeToTitle(field.Name),
				JSONName:      stringutils.SnakeToCamel(field.Name),
				HumanName:     stringutils.SnakeToHuman(field.Name),
				GoType:        field.DataType.GoType(),
				Required:      hasBlankConstraint(field.Constraints),
				IsEmail:       isEmail(field.Name),
			})
		}

		modelFields = append(modelFields, ModelField{
			Name:          field.Name,
			TitleCaseName: stringutils.SnakeToTitle(field.Name),
			CamelCaseName: stringutils.SnakeToCamel(field.Name),
			GoType:        field.DataType.GoType(),
		})
	}

	return searchHandlerTemplateData{
		PackageName:           schema.Name,
		Name:                  schema.Name,
		HumanName:             stringutils.SnakeToHuman(schema.Name),
		ModuleName:            moduleName,
		TitleCaseTableName:    stringutils.SnakeToTitle(schema.Name),
		SingularTitleCaseName: stringutils.SnakeToTitle(strings.TrimSuffix(schema.Name, "s")),
		SingularCamelCaseName: stringutils.SnakeToCamel(strings.TrimSuffix(schema.Name, "s")),
		KebabCaseTableName:    stringutils.SnakeToKebab(schema.Name),
		Fields:                fields,
		ModelFields:           modelFields,
	}
}

func RenderSearchTemplate(moduleName string, schema Table) ([]byte, []byte, error) {
	data := newSearchHandlerTemplateData(moduleName, schema)

	tmpl, err := renderTemplateFile(searchHandlerTemplate, data)
	if err != nil {
		return nil, nil, fmt.Errorf("error rendering search template: %w", err)
	}

	testTmpl, err := renderTemplateFile(searchHandlerTestTemplate, data)
	if err != nil {
		return nil, nil, fmt.Errorf("error rendering search test template: %w", err)
	}

	return tmpl, testTmpl, nil
}
