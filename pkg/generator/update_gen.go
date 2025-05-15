package generator

import (
	"fmt"
	"strings"

	"github.com/gurch101/gowebutils/pkg/stringutils"
)

const updateHandlerTemplate = `package {{.PackageName}}

import (
	"context"
	"net/http"

	"github.com/gurch101/gowebutils/pkg/app"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/parser"
	"github.com/gurch101/gowebutils/pkg/validation"

	{{- range .ForeignKeys}}
	"{{$.ModuleName}}/internal/{{.Table}}"
	{{- end}}
)

type Update{{.SingularTitleCaseName}}Controller struct {
	app *app.App
}

func NewUpdate{{.SingularTitleCaseName}}Controller(app *app.App) *Update{{.SingularTitleCaseName}}Controller {
	return &Update{{.SingularTitleCaseName}}Controller{app: app}
}

type Update{{.SingularTitleCaseName}}Request struct {
{{- range .Fields}}
	{{.TitleCaseName}} *{{.GoType}} ` + "`" + `json:"{{.JSONName}}"` + "`" + `
{{- end}}
}

// Update{{.SingularTitleCaseName}} godoc
//
//	@Summary		Update a {{.HumanName}}
//	@Description	Update a {{.HumanName}} by ID
//	@Tags			{{.HumanName}}s
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int64					true	"{{.SingularTitleCaseName}} ID"
//	@Param			{{.SingularCamelCaseName}}	body		Update{{.SingularTitleCaseName}}Request	true	"Update {{.SingularCamelCaseName}}"
//	@Success		200		{object}	Get{{.SingularTitleCaseName}}ByIDResponse
//	@Failure		400,422,404,500	{object}	httputils.ErrorResponse
//	@Router			/{{.KebabCaseTableName}}/{id} [patch]
func (tc *Update{{.SingularTitleCaseName}}Controller) Update{{.SingularTitleCaseName}}Handler(w http.ResponseWriter, r *http.Request) {
	id, err := parser.ParseIDPathParam(r)

	if err != nil {
		httputils.NotFoundResponse(w, r)

		return
	}

	req, err := httputils.ReadJSON[Update{{.SingularTitleCaseName}}Request](w, r)
	if err != nil {
		httputils.UnprocessableEntityResponse(w, r, err)

		return
	}

	resp, err := Update{{.SingularTitleCaseName}}(r.Context(), tc.app.DB(), id, &req)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)

		return
	}

	err = httputils.WriteJSON(
		w,
		http.StatusOK,
		resp,
		nil)

	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

func Update{{.SingularTitleCaseName}}(
	ctx context.Context,
	db dbutils.DB,
	id int64,
	req *Update{{.SingularTitleCaseName}}Request,
) (*Get{{.SingularTitleCaseName}}ByIDResponse, error) {

	model, err := Get{{.SingularTitleCaseName}}ByID(ctx, db, id)
	if err != nil {
		return nil, err
	}

	{{range .ForeignKeys}}
	if req.{{.TitleCaseFromColumnName}} != nil && *req.{{.TitleCaseFromColumnName}} != model.{{.TitleCaseFromColumnName}} && !{{.Table}}.{{.SingularTitleCaseTableName}}Exists(ctx, db, *req.{{.TitleCaseFromColumnName}}) {
		return nil, Err{{.SingularTitleCaseTableName}}NotFound
	}
	{{- end}}

	{{- range .Fields}}
	model.{{.TitleCaseName}} = validation.Coalesce(req.{{.TitleCaseName}}, model.{{.TitleCaseName}})
	{{- end}}

	{{- if .RequireValidation}}
	v := validation.NewValidator()
	{{- range .Fields}}
	{{- if .IsEmail}}
	v.Email(model.{{.TitleCaseName}}, "{{.JSONName}}", "{{.HumanName}} must be a valid email address")
	{{- else if .Required}}
	v.Required(model.{{.TitleCaseName}}, "{{.JSONName}}", "{{.HumanName}} is required")
	{{- end}}
	{{- end}}

	if v.HasErrors() {
		return nil, v.AsError()
	}
	{{- end}}

	if err := update{{.SingularTitleCaseName}}(ctx, db, model); err != nil {
		return nil, err
	}

	return &Get{{.SingularTitleCaseName}}ByIDResponse{
		{{- range .ModelFields}}
		{{.TitleCaseName}}: model.{{.TitleCaseName}},
		{{- end}}
	}, nil
}

func update{{.SingularTitleCaseName}}(ctx context.Context, db dbutils.DB, model *{{.SingularCamelCaseName}}Model) error {
	return dbutils.UpdateByID(ctx, db, "{{.Name}}", model.ID, model.Version, map[string]any{
		{{- range .Fields}}
		"{{.Name}}": model.{{.TitleCaseName}},
		{{- end}}
	})
}
`

const updateHandlerTestTemplate = `package {{.PackageName}}_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"{{.ModuleName}}/internal/{{.PackageName}}"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestUpdate{{.SingularTitleCaseName}}Handler(t *testing.T) {
	t.Parallel()

	t.Run("successful update", func(t *testing.T) {
		app := testutils.NewTestApp(t)

		defer app.Close()

		ID, _ := {{.PackageName}}.CreateTest{{.SingularTitleCaseName}}(t, app.DB())

		controller := {{.PackageName}}.NewUpdate{{.SingularTitleCaseName}}Controller(app.App)
		app.TestRouter.Patch("/{{.KebabCaseTableName}}/{id}", controller.Update{{.SingularTitleCaseName}}Handler)

		{{if .ForeignKeys}}
		updateReq := {{.PackageName}}.CreateTestUpdate{{.SingularTitleCaseName}}RequestWithValues(t, {{$.PackageName}}.Update{{.SingularTitleCaseName}}Request{
			{{- range .ForeignKeys}}
			{{.TitleCaseFromColumnName}}: testutils.Int64Ptr(1),
			{{- end}}
		})
		{{- else}}
		updateReq := {{.PackageName}}.CreateTestUpdate{{.SingularTitleCaseName}}Request(t)
		{{- end}}

		req := testutils.CreatePatchRequest(t, fmt.Sprintf("/{{.KebabCaseTableName}}/%d", ID), updateReq)
		rr := app.MakeRequest(req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
		}

		var response {{.PackageName}}.Get{{.SingularTitleCaseName}}ByIDResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
		}

		if response.ID != ID {
			t.Errorf("expected ID to be %d, got %d", ID, response.ID)
		}
		{{- range .Fields}}
		if response.{{.TitleCaseName}} != *updateReq.{{.TitleCaseName}} {
			t.Errorf("expected {{.TitleCaseName}} to be %v, got %v", *updateReq.{{.TitleCaseName}}, response.{{.TitleCaseName}})
		}
		{{- end}}
	})

	t.Run("invalid request id", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		defer app.Close()

		controller := {{.PackageName}}.NewUpdate{{.SingularTitleCaseName}}Controller(app.App)
		app.TestRouter.Patch("/{{.KebabCaseTableName}}/{id}", controller.Update{{.SingularTitleCaseName}}Handler)
		req := testutils.CreatePatchRequest(t, "/{{.KebabCaseTableName}}/invalid_id", nil)
		rr := app.MakeRequest(req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d", http.StatusNotFound, rr.Code)
		}
	})

	t.Run("record not found", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		defer app.Close()

		controller := {{.PackageName}}.NewUpdate{{.SingularTitleCaseName}}Controller(app.App)
		app.TestRouter.Patch("/{{.KebabCaseTableName}}/{id}", controller.Update{{.SingularTitleCaseName}}Handler)

		nonExistentID := int64(9999)
		updateReq := {{.PackageName}}.CreateTestUpdate{{.SingularTitleCaseName}}Request(t)

		req := testutils.CreatePatchRequest(t, fmt.Sprintf("/{{.KebabCaseTableName}}/%d", nonExistentID), updateReq)
		rr := app.MakeRequest(req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d", http.StatusNotFound, rr.Code)
		}
	})

	{{- range .ForeignKeys}}
	t.Run("invalid {{.HumanTableName}} foreign key", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		defer app.Close()

		ID, _ := {{$.PackageName}}.CreateTest{{$.SingularTitleCaseName}}(t, app.DB())

		controller := {{$.PackageName}}.NewUpdate{{$.SingularTitleCaseName}}Controller(app.App)
		app.TestRouter.Patch("/{{$.KebabCaseTableName}}/{id}", controller.Update{{$.SingularTitleCaseName}}Handler)

		updateReq := {{$.PackageName}}.CreateTestUpdate{{$.SingularTitleCaseName}}RequestWithValues(t, {{$.PackageName}}.Update{{$.SingularTitleCaseName}}Request{
			{{.TitleCaseFromColumnName}}: testutils.Int64Ptr(2),
		})

		req := testutils.CreatePatchRequest(t, fmt.Sprintf("/{{$.KebabCaseTableName}}/%d", ID), updateReq)
		rr := app.MakeRequest(req)

		testutils.AssertValidationError(t, rr, "{{.JSONName}}", "{{.HumanTableName}} not found")
	});
	{{- end}}

	t.Run("invalid request payload", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		defer app.Close()

		ID, _ := {{.PackageName}}.CreateTest{{.SingularTitleCaseName}}(t, app.DB())

		controller := {{.PackageName}}.NewUpdate{{.SingularTitleCaseName}}Controller(app.App)
		app.TestRouter.Patch("/{{.KebabCaseTableName}}/{id}", controller.Update{{.SingularTitleCaseName}}Handler)

		invalidReq := map[string]interface{}{
			"invalid_field": "value",
		}
		req := testutils.CreatePatchRequest(t, fmt.Sprintf("/{{.KebabCaseTableName}}/%d", ID), invalidReq)
		rr := app.MakeRequest(req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status code %d, got %d", http.StatusUnprocessableEntity, rr.Code)
		}
	})
}
`

func newUpdateHandlerTemplateData(moduleName string, schema Table) updateHandlerTemplateData {
	modelFields := []ModelField{}
	fields := []RequestField{}
	requireValidation := false

	for _, field := range schema.Fields {
		sanitizedName := field.Name

		if strings.HasSuffix(field.Name, "id") {
			sanitizedName = strings.TrimSuffix(field.Name, "id") + "ID"
		}

		if IsRequestField(field) {
			if hasBlankConstraint(field.Constraints) || isEmail(sanitizedName) {
				requireValidation = true
			}

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
			TitleCaseName: stringutils.SnakeToTitle(sanitizedName),
			CamelCaseName: stringutils.SnakeToCamel(sanitizedName),
			GoType:        field.DataType.GoType(),
		})
	}

	return updateHandlerTemplateData{
		PackageName:           schema.Name,
		Name:                  schema.Name,
		KebabCaseTableName:    stringutils.SnakeToKebab(schema.Name),
		ModuleName:            moduleName,
		HumanName:             stringutils.SnakeToHuman(strings.TrimSuffix(schema.Name, "s")),
		SingularTitleCaseName: stringutils.SnakeToTitle(strings.TrimSuffix(schema.Name, "s")),
		SingularCamelCaseName: stringutils.SnakeToCamel(strings.TrimSuffix(schema.Name, "s")),
		RequireValidation:     requireValidation,
		ModelFields:           modelFields,
		Fields:                fields,
		ForeignKeys:           schema.ForeignKeys,
	}
}

func RenderUpdateTemplate(moduleName string, schema Table) ([]byte, []byte, error) {
	data := newUpdateHandlerTemplateData(moduleName, schema)

	tmpl, err := renderTemplateFile(updateHandlerTemplate, data)
	if err != nil {
		return nil, nil, fmt.Errorf("error rendering create template: %w", err)
	}

	testTmpl, err := renderTemplateFile(updateHandlerTestTemplate, data)
	if err != nil {
		return nil, nil, fmt.Errorf("error rendering create test template: %w", err)
	}

	return tmpl, testTmpl, nil
}
