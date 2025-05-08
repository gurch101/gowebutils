package generator

import (
	"fmt"
	"strings"

	"github.com/gurch101/gowebutils/pkg/stringutils"
)

const createHandlerTemplate = `package {{.PackageName}}

import (
	"context"
	{{if .UniqueConstraint}}"errors"{{end}}
	"fmt"
	"net/http"
	{{if .UniqueConstraint}}"strings"{{- end}}

	"github.com/gurch101/gowebutils/pkg/app"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	{{if .RequireValidation}}"github.com/gurch101/gowebutils/pkg/validation"{{end}}

	{{- range .ForeignKeys}}
	"{{$.ModuleName}}/internal/{{.Table}}"
	{{- end}}
)

/* Handler */
type Create{{.SingularTitleCaseName}}Controller struct {
	app *app.App
}

func NewCreate{{.SingularTitleCaseName}}Controller(app *app.App) *Create{{.SingularTitleCaseName}}Controller {
	return &Create{{.SingularTitleCaseName}}Controller{app: app}
}

type Create{{.SingularTitleCaseName}}Request struct {
{{- range .Fields}}
	{{.TitleCaseName}} {{.GoType}} ` + "`" + `json:"{{.JSONName}}"{{.SwaggerTag}}` + "`" + `
{{- end}}
}

type Create{{.SingularTitleCaseName}}Response struct {
	ID int64 ` + "`" + `json:"id"` + "`" + `
}

// Create{{.SingularTitleCaseName}} godoc
//
//	@Summary			Create a {{.HumanName}}
//	@Description	Create a new {{.HumanName}}
//	@Tags					{{.HumanName}}s
//	@Accept				json
//	@Produce			json
//	@Param				{{.SingularCamelCaseName}}	body		Create{{.SingularTitleCaseName}}Request	true	"Create {{.SingularCamelCaseName}}"
//	@Success			201	{object}	Create{{.SingularTitleCaseName}}Response
//	@Header     	201 {string}  Location  "/{{.KebabCaseTableName}}/{id}"
//	@Failure			400,422,404,500	{object}	httputils.ErrorResponse
//	@Router				/{{.KebabCaseTableName}} [post]
func (c *Create{{.SingularTitleCaseName}}Controller) Create{{.SingularTitleCaseName}}Handler(
	w http.ResponseWriter,
	r *http.Request) {
	req, err := httputils.ReadJSON[Create{{.SingularTitleCaseName}}Request](w, r)
	if err != nil {
		httputils.UnprocessableEntityResponse(w, r, err)
		return
	}
	{{if .RequireValidation}}
		v := validation.NewValidator()
		{{- range .Fields}}
			{{- if .IsEmail}}
				v.Email(req.{{.TitleCaseName}}, "{{.JSONName}}", "{{.HumanName}} must be a valid email address")
			{{- else if .Required}}
				{{- if eq .GoType "string"}}
				v.Required(req.{{.TitleCaseName}}, "{{.JSONName}}", "{{.HumanName}} is required")
				{{- else}}
				v.Check(req.{{.TitleCaseName}} > 0, "{{.JSONName}}", "{{.HumanName}} is required")
				{{- end}}
			{{- end}}
		{{- end}}

		if v.HasErrors() {
			httputils.FailedValidationResponse(w, r, v.Errors)
			return
		}
	{{- end}}

	id, err := Create{{.SingularTitleCaseName}}(r.Context(), c.app.DB(), &req)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/{{.KebabCaseTableName}}/%d", *id))

	err = httputils.WriteJSON(w, http.StatusCreated, Create{{.SingularTitleCaseName}}Response{ID: *id}, headers)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

/* Service */
func Create{{.SingularTitleCaseName}}(
	ctx context.Context,
	db dbutils.DB,
	req *Create{{.SingularTitleCaseName}}Request) (*int64, error) {

	{{range .ForeignKeys}}
	if !{{.Table}}.{{.SingularTitleCaseTableName}}Exists(ctx, db, req.{{.TitleCaseFromColumnName}}) {
		return nil, Err{{.SingularTitleCaseTableName}}NotFound
	}
	{{- end}}

	model := newCreate{{.SingularTitleCaseName}}Model(
		{{- range .Fields}}
		req.{{.TitleCaseName}},
		{{- end}}
	)

	{{if .UniqueConstraint}}
	id, err := insert{{.SingularTitleCaseName}}(ctx, db, model)

	if err != nil {
		if errors.Is(err, dbutils.ErrUniqueConstraint) {
			{{- range .UniqueFields}}
			if strings.Contains(err.Error(), "{{.Name}}") {
				return nil, Err{{.TitleCaseName}}AlreadyExists
			}
			{{- end}}
		}

		return nil, err
	}

	return id, nil
	{{- else}}
	return insert{{.SingularTitleCaseName}}(ctx, db, model)
	{{- end}}
}

/* Repository */
func insert{{.SingularTitleCaseName}}(
	ctx context.Context,
	db dbutils.DB,
	model *{{.SingularCamelCaseName}}Model) (*int64, error) {

	return dbutils.Insert(ctx, db, "{{.PackageName}}", map[string]any{
		{{- range .Fields}}
		"{{.Name}}": model.{{.TitleCaseName}},
		{{- end}}
	})
}
`

const createHandlerTestTemplate = `package {{.PackageName}}_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"{{.ModuleName}}/internal/{{.PackageName}}"
	{{- range .ForeignKeys}}
	"{{$.ModuleName}}/internal/{{.Table}}"
	{{- end}}
	{{- if or (.RequireValidation) (.ForeignKeys)}}
	"github.com/gurch101/gowebutils/pkg/collectionutils"
	"github.com/gurch101/gowebutils/pkg/validation"
	{{- end}}
	{{if .ForeignKeys}}"github.com/gurch101/gowebutils/pkg/httputils"{{end}}
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestCreate{{.SingularTitleCaseName}}(t *testing.T) {
t.Parallel()

	t.Run("successful create", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		defer app.Close()

		controller := {{.PackageName}}.NewCreate{{.SingularTitleCaseName}}Controller(app.App)
		app.TestRouter.Post("/{{.KebabCaseTableName}}", controller.Create{{.SingularTitleCaseName}}Handler)

		{{- range .ForeignKeys}}
		{{.SingularCamelCaseTableName}}ID, _ := {{.Table}}.CreateTest{{.SingularTitleCaseTableName}}(t, app.DB())
		{{- end}}
		body := {{.PackageName}}.CreateTest{{.SingularTitleCaseName}}Request(t)
		{{- range .ForeignKeys}}
		body.{{.SingularTitleCaseTableName}}ID = {{.SingularCamelCaseTableName}}ID
		{{- end}}
		req := testutils.CreatePostRequest(t, "/{{.KebabCaseTableName}}", body)
		rr := app.MakeRequest(req)

		if rr.Code != http.StatusCreated {
			t.Errorf("expected status code %d, got %d", http.StatusCreated, rr.Code)
		}

		var response {{.PackageName}}.Create{{.SingularTitleCaseName}}Response
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
		}

		if response.ID <= 0 {
			t.Errorf("expected ID to be positive, got %d", response.ID)
		}

		location := rr.Header().Get("Location")
		if location == "" {
			t.Errorf("expected Location header to be set")
		}

		if location != fmt.Sprintf("/{{.KebabCaseTableName}}/%d", response.ID) {
			t.Errorf("expected Location header to be %s, got %s", fmt.Sprintf("/{{.KebabCaseTableName}}/%d", response.ID), location)
		}

		{{range .Fields}}
		var {{.JSONName}} {{.GoType}}
		{{- end}}

		err = app.DB().QueryRowContext(context.Background(), fmt.Sprintf("SELECT {{ range $i, $field := .Fields }} {{if eq $i 0}}{{.Name}}{{else}},{{.Name}}{{end}} {{end}} FROM {{.Name}} WHERE id = %d", response.ID)).Scan(
			{{- range .Fields}}
			&{{.JSONName}},
			{{- end}}
		)
		if err != nil {
			t.Fatal(err)
		}

		{{- range .Fields}}
		if {{.JSONName}} != body.{{.TitleCaseName}} {
			t.Errorf("expected {{.JSONName}} to be %v, got %v", body.{{.TitleCaseName}}, {{.JSONName}})
		}
		{{- end}}
	})

	t.Run("invalid request body", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		defer app.Close()

		controller := {{.PackageName}}.NewCreate{{.SingularTitleCaseName}}Controller(app.App)
		app.TestRouter.Post("/{{.KebabCaseTableName}}", controller.Create{{.SingularTitleCaseName}}Handler)

		payload := map[string]interface{}{
			"invalid": "",
		}
		req := testutils.CreatePostRequest(t, "/{{.KebabCaseTableName}}", payload)
		rr := app.MakeRequest(req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Errorf("Expected status code 422 Unprocessable Entity, got %d", rr.Code)
		}
	})

	{{if .RequireValidation}}
	t.Run("failed request validation", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		defer app.Close()

		controller := {{.PackageName}}.NewCreate{{.SingularTitleCaseName}}Controller(app.App)
		app.TestRouter.Post("/{{.KebabCaseTableName}}", controller.Create{{.SingularTitleCaseName}}Handler)

		body := {{.PackageName}}.Create{{.SingularTitleCaseName}}Request{
			{{- range .Fields}}
				{{- if .IsEmail}}
				{{.TitleCaseName}}: "invalidemail",
				{{- else if .Required}}
						{{- if or (eq .GoType "int") (eq .GoType "int64")}}
						{{.TitleCaseName}}: 0,
						{{- else}}
						{{.TitleCaseName}}: "",
						{{- end}}
				{{- end}}
			{{- end}}
		}

		req := testutils.CreatePostRequest(t, "/{{.KebabCaseTableName}}", body)
		rr := app.MakeRequest(req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code 400 Bad Request, got %d", rr.Code)
		}

		// Verify error message
		var errorResponse testutils.ValidationErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)

		if err != nil {
			t.Fatal(err)
		}

		if len(errorResponse.Errors) == 0 {
			t.Error("Expected validation errors, got none")
		}

		var ok bool
		{{range .Fields}}
			{{- if .IsEmail}}
			ok = collectionutils.Contains(errorResponse.Errors, func(e validation.Error) bool {
				return e.Field == "{{.JSONName}}" && e.Message == "{{.HumanName}} must be a valid email address"
			})

			if !ok {
				t.Errorf("Expected error message for {{.JSONName}}, but got none")
			}
			{{- else if .Required}}
			ok = collectionutils.Contains(errorResponse.Errors, func(e validation.Error) bool {
				return e.Field == "{{.JSONName}}" && e.Message == "{{.HumanName}} is required"
			})

			if !ok {
				t.Errorf("Expected error message for {{.JSONName}}, but got none")
			}
			{{- end}}
		{{- end}}
	})
	{{- end}}

	{{if .UniqueConstraint}}
	t.Run("failed unique constraints", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		defer app.Close()

		controller := {{.PackageName}}.NewCreate{{.SingularTitleCaseName}}Controller(app.App)
		app.TestRouter.Post("/{{.KebabCaseTableName}}", controller.Create{{.SingularTitleCaseName}}Handler)

		_, payload := {{.PackageName}}.CreateTest{{.SingularTitleCaseName}}(t, app.DB())
		{{.PackageName}}.Create{{.SingularTitleCaseName}}(context.Background(), app.DB(), &payload)

		req := testutils.CreatePostRequest(t, "/{{.KebabCaseTableName}}", payload)
		rr := app.MakeRequest(req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code 400 Bad Request, got %d", rr.Code)
		}

		var errorResponse testutils.ValidationErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)

		if err != nil {
			t.Fatal(err)
		}

		if len(errorResponse.Errors) == 0 {
			t.Error("Expected validation errors, got none")
		}

		var ok bool
		{{range .UniqueFields}}
		ok = collectionutils.Contains(errorResponse.Errors, func(e validation.Error) bool {
			return e.Field == "{{.JSONName}}" && e.Message == "{{.HumanName}} already exists"
		})

		if !ok {
			t.Errorf("Expected error message for {{.JSONName}}, but got none")
		}
		{{- end}}
	})
	{{- end}}

	{{range .ForeignKeys}}
	t.Run("failed {{.HumanTableName}} foreign key constraint", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		defer app.Close()

		controller := {{$.PackageName}}.NewCreate{{$.SingularTitleCaseName}}Controller(app.App)
		app.TestRouter.Post("/{{$.KebabCaseTableName}}", controller.Create{{$.SingularTitleCaseName}}Handler)

		_, payload := {{$.PackageName}}.CreateTest{{$.SingularTitleCaseName}}(t, app.DB())
		payload.{{.TitleCaseFromColumnName}} = 100

		req := testutils.CreatePostRequest(t, "/{{$.KebabCaseTableName}}", payload)
		rr := app.MakeRequest(req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code 400 Bad Request, got %d", rr.Code)
		}

		var errorResponse httputils.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
		if err != nil {
			t.Fatal(err)
		}

		ok := collectionutils.Contains(errorResponse.Errors, func(e validation.Error) bool {
			return e.Field == "{{.JSONName}}" && e.Message == "{{.HumanTableName}} not found"
		})

		if !ok {
			t.Errorf("Expected error message for {{.JSONName}}, but got none")
		}
	})
	{{- end}}
}
`

func hasBlankConstraint(constraints []string) bool {
	for _, constraint := range constraints {
		if strings.Contains(constraint, "CHECK") && strings.Contains(constraint, "<> ''") {
			return true
		}
	}

	return false
}

func hasUniqueConstraint(constraints []string) bool {
	for _, constraint := range constraints {
		if strings.Contains(constraint, "UNIQUE") {
			return true
		}
	}

	return false
}

func isNotNullForeignKey(schema Table, field Field) bool {
	isNotNull := false

	for _, constraint := range field.Constraints {
		if strings.Contains(constraint, "NOT NULL") {
			isNotNull = true

			break
		}
	}

	if isNotNull {
		for _, fk := range schema.ForeignKeys {
			if fk.FromColumn == field.Name {
				return true
			}
		}
	}

	return false
}

func isEmail(fieldName string) bool {
	return strings.Contains(fieldName, "email")
}

func getSanitizedName(name string) string {
	sanitizedName := name
	if strings.HasSuffix(sanitizedName, "id") {
		sanitizedName = strings.TrimSuffix(sanitizedName, "id") + "ID"
	}

	return sanitizedName
}

func newCreateHandlerTemplateData(moduleName string, schema Table) createHandlerTemplateData {
	fields := []RequestField{}
	uniqueFields := []UniqueField{}
	modelFields := []ModelField{}
	includeUniqueConstraint := false
	requireValidation := false

	for _, field := range schema.Fields {
		sanitizedName := getSanitizedName(field.Name)

		if IsRequestField(field) {
			if hasUniqueConstraint(field.Constraints) {
				includeUniqueConstraint = true
				requireValidation = true

				uniqueFields = append(uniqueFields, newUniqueField(field))
			}

			required := hasBlankConstraint(field.Constraints) || isNotNullForeignKey(schema, field)
			email := isEmail(sanitizedName)

			if required || email {
				requireValidation = true
			}

			fields = append(fields, RequestField{
				Name:          field.Name,
				TitleCaseName: stringutils.SnakeToTitle(sanitizedName),
				JSONName:      stringutils.SnakeToCamel(sanitizedName),
				HumanName:     stringutils.SnakeToHuman(sanitizedName),
				GoType:        field.DataType.GoType(),
				Required:      required,
				IsEmail:       email,
			})
		}

		modelFields = append(modelFields, newModelField(sanitizedName, field))
	}

	return createHandlerTemplateData{
		PackageName:           schema.Name,
		Name:                  schema.Name,
		ModuleName:            moduleName,
		HumanName:             stringutils.SnakeToHuman(strings.TrimSuffix(schema.Name, "s")),
		UniqueConstraint:      includeUniqueConstraint,
		RequireValidation:     requireValidation,
		TitleCaseTableName:    stringutils.SnakeToTitle(schema.Name),
		SingularTitleCaseName: stringutils.SnakeToTitle(strings.TrimSuffix(schema.Name, "s")),
		SingularCamelCaseName: stringutils.SnakeToCamel(strings.TrimSuffix(schema.Name, "s")),
		KebabCaseTableName:    strings.ToLower(stringutils.SnakeToKebab(schema.Name)),
		UniqueFields:          uniqueFields,
		Fields:                fields,
		ModelFields:           modelFields,
		ForeignKeys:           schema.ForeignKeys,
	}
}

func RenderCreateTemplate(moduleName string, schema Table) ([]byte, []byte, error) {
	data := newCreateHandlerTemplateData(moduleName, schema)

	createTemplate, err := renderTemplateFile(createHandlerTemplate, data)
	if err != nil {
		return nil, nil, fmt.Errorf("error rendering create template: %w", err)
	}

	createTestTemplate, err := renderTemplateFile(createHandlerTestTemplate, data)
	if err != nil {
		return nil, nil, fmt.Errorf("error rendering create test template: %w", err)
	}

	return createTemplate, createTestTemplate, nil
}
