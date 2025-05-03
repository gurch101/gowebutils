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
	"github.com/gurch101/gowebutils/pkg/stringutils"
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
	{{- range .ModelFields}}
	{{.TitleCaseName}} {{.GoType}} ` + "`" + `json:"{{.CamelCaseName}}"` + "`" + `
	{{- end}}
}

func (tc *Search{{.SingularTitleCaseName}}Controller) Search{{.SingularTitleCaseName}}Handler(w http.ResponseWriter, r *http.Request) {
	queryString := r.URL.Query()
	request := &Search{{.SingularTitleCaseName}}Request{
		{{- range .Fields}}
		{{.TitleCaseName}}: parser.ParseQS{{if eq .GoType "bool"}}Bool{{else}}String{{end}}(queryString, "{{.JSONName}}", nil),
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

	response, pagination, err := Search{{.TitleCaseTableName}}(r.Context(), tc.app.DB(), request)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)
		return
	}

	filteredResponse, err := parser.StructsToFilteredMaps(response, request.Fields)

	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
		return
	}

	err = httputils.WriteJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"metadata": pagination,
			"{{.Name}}": filteredResponse,
		}, nil,
	)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}


func Search{{.TitleCaseTableName}}(
	ctx context.Context,
	db dbutils.DB,
	search{{.SingularTitleCaseName}}Request *Search{{.SingularTitleCaseName}}Request,
) ([]*Search{{.SingularTitleCaseName}}Response, parser.PaginationMetadata, error) {
	models, pagination, err := find{{.TitleCaseTableName}}(ctx, db, search{{.SingularTitleCaseName}}Request)
	if err != nil {
		return nil, pagination, err
	}

	response := make([]*Search{{.SingularTitleCaseName}}Response, 0)

	for _, model := range models {
		{{.SingularCamelCaseName}} := &Search{{.SingularTitleCaseName}}Response{
		{{- range .ModelFields}}
			{{.TitleCaseName}}: model.{{.TitleCaseName}},
		{{- end}}
		}
		response = append(response, {{.SingularCamelCaseName}})
	}
	return response, pagination, nil
}

func find{{.TitleCaseTableName}}(
	ctx context.Context,
	db dbutils.DB,
	request *Search{{.SingularTitleCaseName}}Request) ([]{{.SingularCamelCaseName}}Model, parser.PaginationMetadata, error) {
	var models []{{.SingularCamelCaseName}}Model
	var totalRecords int

	dbFields := make([]string, len(request.Fields) + 1)
	dbFields[0] = "count(*) over()"

	for i, field := range request.Fields {
		dbFields[i + 1] = stringutils.CamelToSnake(field)
	}

	err := dbutils.NewQueryBuilder(db).
		Select(
			dbFields...
		).
		From("{{.Name}}").
		{{range $i, $field := .Fields}}
			{{if eq $i 0}}
			Where("{{$field.Name}} = ?", request.{{$field.TitleCaseName}}).
			{{else}}
			AndWhere("{{$field.Name}} = ?", request.{{$field.TitleCaseName}}).
			{{end}}
		{{end}}
		OrderBy(request.Sort).
		Page(request.Page, request.PageSize).
		QueryContext(ctx, func(rows *sql.Rows) error {
			var model {{.SingularCamelCaseName}}Model

			fieldsToBindTo := make([]interface{}, len(dbFields))
			fieldsToBindTo[0] = &totalRecords
			for i, field := range dbFields[1:] {
				fieldsToBindTo[i + 1] = model.Field(field)
			}

			err := rows.Scan(
				fieldsToBindTo...,
			)

			if err != nil {
				return err
			}

			models = append(models, model)
			return nil
		})

	if err != nil {
		return nil, parser.PaginationMetadata{}, dbutils.WrapDBError(err)
	}

	metadata := parser.ParsePaginationMetadata(totalRecords, request.Page, request.PageSize)
	return models, metadata, nil
}
`

const searchHandlerTestTemplate = `package {{.PackageName}}_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/gurch101/gowebgentest/internal/{{.PackageName}}"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestSearch{{.SingularTitleCaseName}}(t *testing.T) {
    t.Parallel()

    t.Run("successful search", func(t *testing.T) {
        app := testutils.NewTestApp(t)
        defer app.Close()

        body := {{.PackageName}}.Create{{.SingularTitleCaseName}}Request{
            {{- range .Fields}}
                {{- if .IsEmail}}
                {{.TitleCaseName}}: "{{.JSONName}}@example.com",
                {{- else}}
                {{.TitleCaseName}}: "{{.JSONName}}",
                {{- end}}
            {{- end}}
        }

        _, err := {{.PackageName}}.Create{{.SingularTitleCaseName}}(context.Background(), app.DB(), &body)
        if err != nil {
            t.Fatal(err)
        }

        controller := {{.PackageName}}.NewSearch{{.SingularTitleCaseName}}Controller(app.App)
        app.TestRouter.Get("/{{.KebabCaseTableName}}", controller.Search{{.SingularTitleCaseName}}Handler)

        req := testutils.CreateGetRequest(t, "/{{.KebabCaseTableName}}")
        rr := app.MakeRequest(req)

        if rr.Code != http.StatusOK {
            t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
        }

        var response struct {
            Data []{{.PackageName}}.Search{{.SingularTitleCaseName}}Response ` + "`" + `json:"{{.Name}}"` + "`" + `
        }
        err = json.Unmarshal(rr.Body.Bytes(), &response)
        if err != nil {
            t.Fatal(err)
        }

        if len(response.Data) != 1 {
            t.Errorf("expected 1 user, got %d", len(response.Data))
        }

				actualRecord, err := {{.PackageName}}.Get{{.SingularTitleCaseName}}ByID(context.Background(), app.DB(), response.Data[0].ID)
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

        if rr.Code != http.StatusBadRequest {
            t.Errorf("expected status code %d, got %d", http.StatusBadRequest, rr.Code)
        }

        var response map[string]interface{}
        err := json.Unmarshal(rr.Body.Bytes(), &response)
        if err != nil {
            t.Fatal(err)
        }

        if _, ok := response["errors"]; !ok {
            t.Error("expected validation errors in response")
        }
    })

    t.Run("bad field parameter", func(t *testing.T) {
        app := testutils.NewTestApp(t)
        defer app.Close()

        controller := {{.PackageName}}.NewSearch{{.SingularTitleCaseName}}Controller(app.App)
        app.TestRouter.Get("/{{.KebabCaseTableName}}", controller.Search{{.SingularTitleCaseName}}Handler)

        req := testutils.CreateGetRequest(t, "/{{.KebabCaseTableName}}?fields=invalidField")
        rr := app.MakeRequest(req)

        if rr.Code != http.StatusBadRequest {
            t.Errorf("expected status code %d, got %d", http.StatusBadRequest, rr.Code)
        }

        var response map[string]interface{}
        err := json.Unmarshal(rr.Body.Bytes(), &response)
        if err != nil {
            t.Fatal(err)
        }

        if _, ok := response["errors"]; !ok {
            t.Error("expected validation errors in response")
        }
    })

		t.Run("single field", func(t *testing.T) {
			app := testutils.NewTestApp(t)
			defer app.Close()

			body := {{.PackageName}}.Create{{.SingularTitleCaseName}}Request{
					{{- range .Fields}}
						{{- if .IsEmail}}
						{{.TitleCaseName}}: "{{.JSONName}}@example.com",
						{{- else}}
						{{.TitleCaseName}}: "{{.JSONName}}",
						{{- end}}
					{{- end}}
			}
			_, err := {{.PackageName}}.Create{{.SingularTitleCaseName}}(context.Background(), app.DB(), &body)
			if err != nil {
					t.Fatal(err)
			}

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
		sanitizedName := field.Name
		if strings.HasSuffix(field.Name, "id") {
			sanitizedName = strings.TrimSuffix(field.Name, "id")
		}

		if field.Name != "id" && field.Name != "version" && field.Name != "created_at" && field.Name != "updated_at" {
			fields = append(fields, RequestField{
				Name:          field.Name,
				TitleCaseName: stringutils.SnakeToTitle(sanitizedName),
				JSONName:      stringutils.SnakeToCamel(sanitizedName),
				HumanName:     stringutils.SnakeToHuman(sanitizedName),
				GoType:        field.DataType.GoType(),
				Required:      hasBlankConstraint(field.Constraints),
				IsEmail:       isEmail(sanitizedName),
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
		ModuleName:            moduleName,
		TitleCaseTableName:    stringutils.SnakeToTitle(schema.Name),
		SingularTitleCaseName: stringutils.SnakeToTitle(strings.TrimSuffix(schema.Name, "s")),
		SingularCamelCaseName: strings.ToLower(stringutils.SnakeToCamel(strings.TrimSuffix(schema.Name, "s"))),
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
