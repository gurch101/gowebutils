package generator

import (
	"fmt"
	"strings"

	"github.com/gurch101/gowebutils/pkg/stringutils"
)

const getHandlerTemplate = `package {{.PackageName}}

import (
	"context"
	"net/http"
	"time"

	"github.com/gurch101/gowebutils/pkg/app"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/parser"
)

type Get{{.SingularTitleCaseName}}ByIDController struct {
	app *app.App
}

func NewGet{{.SingularTitleCaseName}}ByIDController(app *app.App) *Get{{.SingularTitleCaseName}}ByIDController {
	return &Get{{.SingularTitleCaseName}}ByIDController{app: app}
}

type Get{{.SingularTitleCaseName}}ByIDResponse struct {
	{{- range .ModelFields}}
	{{.TitleCaseName}} {{.GoType}} ` + "`" + `json:"{{.JSONName}}"` + "`" + `
	{{- end}}
}

// Get{{.SingularTitleCaseName}} godoc
//
//	@Summary		Get a {{.HumanName}}
//	@Description	get {{.HumanName}} by ID
//	@Tags			{{.HumanName}}s
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int64	true	"{{.SingularCamelCaseName}} ID"
//	@Success		200	{object}	Get{{.SingularTitleCaseName}}ByIDResponse
//	@Failure		400,422,404,500	{object}	httputils.ErrorResponse
//	@Router			/{{.KebabCaseTableName}}/{id} [get]
func (tc *Get{{.SingularTitleCaseName}}ByIDController) Get{{.SingularTitleCaseName}}ByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parser.ParseIDPathParam(r)

	if err != nil {
		httputils.NotFoundResponse(w, r)
		return
	}

	model, err := Get{{.SingularTitleCaseName}}ByID(r.Context(), tc.app.DB(), id)

	if err != nil {
		httputils.HandleErrorResponse(w, r, err)
		return
	}

	err = httputils.WriteJSON(w, http.StatusOK, &Get{{.SingularTitleCaseName}}ByIDResponse{
	{{- range .ModelFields}}
	{{.TitleCaseName}}: model.{{.TitleCaseName}},
	{{- end}}
	}, nil)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

func Get{{.SingularTitleCaseName}}ByID(ctx context.Context, db dbutils.DB, {{.SingularCamelCaseName}}ID int64) (*{{.SingularCamelCaseName}}Model, error) {
	var model {{.SingularCamelCaseName}}Model

	err := dbutils.GetByID(ctx, db, "{{.Name}}", {{.SingularCamelCaseName}}ID, map[string]any{
	{{- range .ModelFields}}
	"{{.Name}}": &model.{{.TitleCaseName}},
	{{- end}}
	})
	if err != nil {
		return nil, dbutils.WrapDBError(err)
	}
	return &model, nil
}
`

const getHandlerTestTemplate = `package {{.PackageName}}_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"{{.ModuleName}}/internal/{{.PackageName}}"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestGet{{.SingularTitleCaseName}}ByID(t *testing.T) {
	t.Parallel()

	t.Run("successful get by id", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		defer app.Close()

		ID, createReq := {{.PackageName}}.CreateTest{{.SingularTitleCaseName}}(t, app.DB())

		controller := {{.PackageName}}.NewGet{{.SingularTitleCaseName}}ByIDController(app.App)
		app.TestRouter.Get("/{{.KebabCaseTableName}}/{id}", controller.Get{{.SingularTitleCaseName}}ByIDHandler)

		req := testutils.CreateGetRequest(t, fmt.Sprintf("/{{.KebabCaseTableName}}/%d", ID))
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
		{{- range .CreateFields}}
		if response.{{.TitleCaseName}} != createReq.{{.TitleCaseName}} {
			t.Errorf("expected {{.JSONName}} to be %v, got %v", createReq.{{.TitleCaseName}}, response.{{.TitleCaseName}})
		}
		{{- end}}
		{{- if .HasCreatedAt}}
		if response.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set")
		}
		{{- end}}
		{{- if .HasUpdatedAt}}
		if response.UpdatedAt.IsZero() {
			t.Error("expected UpdatedAt to be set")
		}
		{{- end}}
	})

	t.Run("record not found", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		defer app.Close()

		controller := {{.PackageName}}.NewGet{{.SingularTitleCaseName}}ByIDController(app.App)
		app.TestRouter.Get("/{{.KebabCaseTableName}}/{id}", controller.Get{{.SingularTitleCaseName}}ByIDHandler)

		nonExistentID := int64(9999)
		req := testutils.CreateGetRequest(t, fmt.Sprintf("/{{.KebabCaseTableName}}/%d", nonExistentID))
		rr := app.MakeRequest(req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d", http.StatusNotFound, rr.Code)
		}
	})

	t.Run("invalid ID format", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		defer app.Close()

		controller := {{.PackageName}}.NewGet{{.SingularTitleCaseName}}ByIDController(app.App)
		app.TestRouter.Get("/{{.KebabCaseTableName}}/{id}", controller.Get{{.SingularTitleCaseName}}ByIDHandler)

		req := testutils.CreateGetRequest(t, "/{{.KebabCaseTableName}}/invalid")
		rr := app.MakeRequest(req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d", http.StatusNotFound, rr.Code)
		}
	})
}
`

func newGetOneHandlerTemplateData(moduleName string, schema Table) getHandlerTemplateData {
	modelFields := []ModelField{}
	createFields := []RequestField{}
	hasCreatedAt := false
	hasUpdatedAt := false

	for _, field := range schema.Fields {
		sanitizedName := field.Name
		if strings.HasSuffix(field.Name, "id") {
			sanitizedName = strings.TrimSuffix(field.Name, "id") + "ID"
		}

		if field.Name == "created_at" {
			hasCreatedAt = true
		}

		if field.Name == "updated_at" {
			hasUpdatedAt = true
		}

		modelFields = append(modelFields, ModelField{
			Name:          field.Name,
			TitleCaseName: stringutils.SnakeToTitle(sanitizedName),
			CamelCaseName: stringutils.SnakeToCamel(sanitizedName),
			JSONName:      stringutils.SnakeToCamel(sanitizedName),
			GoType:        field.DataType.GoType(),
		})

		if IsRequestField(field) {
			required := hasBlankConstraint(field.Constraints)
			email := isEmail(sanitizedName)

			createFields = append(createFields, RequestField{
				Name:          field.Name,
				TitleCaseName: stringutils.SnakeToTitle(sanitizedName),
				JSONName:      stringutils.SnakeToCamel(sanitizedName),
				HumanName:     stringutils.SnakeToHuman(sanitizedName),
				GoType:        field.DataType.GoType(),
				Required:      required,
				IsEmail:       email,
			})
		}
	}

	return getHandlerTemplateData{
		PackageName:           schema.Name,
		Name:                  schema.Name,
		ModuleName:            moduleName,
		HumanName:             stringutils.SnakeToHuman(strings.TrimSuffix(schema.Name, "s")),
		SingularTitleCaseName: stringutils.SnakeToTitle(strings.TrimSuffix(schema.Name, "s")),
		SingularCamelCaseName: stringutils.SnakeToCamel(strings.TrimSuffix(schema.Name, "s")),
		KebabCaseTableName:    stringutils.SnakeToKebab(schema.Name),
		ModelFields:           modelFields,
		CreateFields:          createFields,
		HasCreatedAt:          hasCreatedAt,
		HasUpdatedAt:          hasUpdatedAt,
	}
}

func RenderGetOneTemplate(moduleName string, schema Table) ([]byte, []byte, error) {
	data := newGetOneHandlerTemplateData(moduleName, schema)

	tmpl, err := renderTemplateFile(getHandlerTemplate, data)
	if err != nil {
		return nil, nil, fmt.Errorf("error rendering create template: %w", err)
	}

	testTmpl, err := renderTemplateFile(getHandlerTestTemplate, data)
	if err != nil {
		return nil, nil, fmt.Errorf("error rendering create test template: %w", err)
	}

	return tmpl, testTmpl, nil
}
