package generator

import (
	"fmt"
	"strings"

	"github.com/gurch101/gowebutils/pkg/stringutils"
)

const createHandlerTemplate = `package {{.PackageName}}

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	{{if .UniqueConstraint}}"strings"{{end}}

	"github.com/gurch101/gowebutils/pkg/app"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	{{if .RequireValidation}}"github.com/gurch101/gowebutils/pkg/validation"{{end}}
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
	{{.TitleCaseName}} {{.GoType}} ` + "`" + `json:"{{.JSONName}}"` + "`" + `
{{- end}}
}

type Create{{.SingularTitleCaseName}}Response struct {
	ID int64 ` + "`" + `json:"id"` + "`" + `
}

// Create{{.SingularTitleCaseName}} godoc
//
//	@Summary		Create a {{.SingularCamelCaseName}}
//	@Description	Create a new {{.SingularCamelCaseName}}
//	@Tags			{{.Name}}
//	@Accept			json
//	@Produce		json
//	@Param			{{.SingularCamelCaseName}}	body		Create{{.SingularTitleCaseName}}Request	true	"Create {{.SingularCamelCaseName}}"}"
//	@Success		201	{object}	CreateUserResponse
//	@Header     201 {string}  Location  "/{{.KebabCaseTableName}}/{id}"
//	@Failure		400,422,404,500	{object}	httputils.ErrorResponse
//	@Router			/{{.KebabCaseTableName}} [post]
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
				v.Email(req.{{.TitleCaseName}}, "{{.JSONName}}", "{{.HumanName}} is required")
			{{- else if .Required}}
				v.Required(req.{{.TitleCaseName}}, "{{.JSONName}}", "{{.HumanName}} is required")
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
{{- range .UniqueFields}}
var Err{{.TitleCaseName}}AlreadyExists = validation.Error{
	Field:   "{{.JSONName}}",
	Message: "{{.HumanName}} already exists",
}
{{- end}}

func Create{{.SingularTitleCaseName}}(
	ctx context.Context,
	db dbutils.DB,
	req *Create{{.SingularTitleCaseName}}Request) (*int64, error) {

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

	"{{.ModuleName}}/internal/users"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestCreate{{.SingularTitleCaseName}}(t *testing.T) {
t.Parallel()

	t.Run("successful create", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		defer app.Close()

		controller := {{.PackageName}}.NewCreate{{.SingularTitleCaseName}}Controller(app.App)
		app.TestRouter.Post("/{{.KebabCaseTableName}}", controller.Create{{.SingularTitleCaseName}}Handler)

		body := {{.PackageName}}.Create{{.SingularTitleCaseName}}Request{
			{{- range .Fields}}
				{{- if .IsEmail}}
				{{.TitleCaseName}}: "{{.JSONName}}@example.com",
				{{- else}}
				{{.TitleCaseName}}: "{{.JSONName}}",
				{{- end}}
			{{- end}}
		}

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
			t.Errorf("expected {{.JSONName}} to be %s, got %s", body.{{.TitleCaseName}}, {{.JSONName}})
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
				{{.TitleCaseName}}: "",
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

		{{$idx := 0}}
		{{range .Fields}}
			{{- if .IsEmail}}
			if errorResponse.Errors[{{$idx}}].Field != "{{.JSONName}}" {
				t.Errorf("Expected error field to be '{{.JSONName}}', got %s", errorResponse.Errors[{{$idx}}].Field)
			}

			if errorResponse.Errors[{{$idx}}].Message != "{{.HumanName}} must be a valid email address" {
				t.Errorf("Expected error message to be '{{.HumanName}} must be a valid email address', got %s", errorResponse.Errors[{{$idx}}].Message)
			}
			{{$idx = incr $idx}}
			{{- else if .Required}}
			if errorResponse.Errors[{{$idx}}].Field != "{{.JSONName}}" {
				t.Errorf("Expected error field to be '{{.JSONName}}', got %s", errorResponse.Errors[{{$idx}}].Field)
			}

			if errorResponse.Errors[{{$idx}}].Message != "{{.HumanName}} is required" {
				t.Errorf("Expected error message to be '{{.HumanName}} is required', got %s", errorResponse.Errors[{{$idx}}].Message)
			}
			{{$idx = incr $idx}}
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

		{{.PackageName}}.Create{{.SingularTitleCaseName}}(context.Background(), app.DB(), &{{.PackageName}}.Create{{.SingularTitleCaseName}}Request{
			{{- range .Fields}}
				{{- if .IsEmail}}
				{{.TitleCaseName}}: "{{.JSONName}}@example.com",
				{{- else if .Required}}
				{{.TitleCaseName}}: "{{.JSONName}}",
				{{- end}}
			{{- end}}
		})

		payload := {{.PackageName}}.Create{{.SingularTitleCaseName}}Request{
			{{- range .Fields}}
				{{- if .IsEmail}}
				{{.TitleCaseName}}: "{{.JSONName}}@example.com",
				{{- else if .Required}}
				{{.TitleCaseName}}: "{{.JSONName}}",
				{{- end}}
			{{- end}}
		}

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

		{{$idx := 0}}
		{{range .UniqueFields}}
		if errorResponse.Errors[{{$idx}}].Field != "{{.JSONName}}" {
			t.Errorf("Expected error field to be '{{.JSONName}}', got %s", errorResponse.Errors[{{$idx}}].Field)
		}

		if errorResponse.Errors[{{$idx}}].Message != "{{.HumanName}} already exists" {
			t.Errorf("Expected error message to be '{{.HumanName}} already exists', got %s", errorResponse.Errors[{{$idx}}].Message)
		}
		{{$idx = incr $idx}}
		{{- end}}
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

		if field.Name != "id" && field.Name != "version" && field.Name != "created_at" && field.Name != "updated_at" {
			if hasUniqueConstraint(field.Constraints) {
				includeUniqueConstraint = true
				requireValidation = true
				uniqueFields = append(uniqueFields, newUniqueField(field))
			}

			required := hasBlankConstraint(field.Constraints)
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
		UniqueConstraint:      includeUniqueConstraint,
		RequireValidation:     requireValidation,
		TitleCaseTableName:    stringutils.SnakeToTitle(schema.Name),
		SingularTitleCaseName: stringutils.SnakeToTitle(strings.TrimSuffix(schema.Name, "s")),
		SingularCamelCaseName: strings.ToLower(stringutils.SnakeToCamel(strings.TrimSuffix(schema.Name, "s"))),
		KebabCaseTableName:    strings.ToLower(stringutils.SnakeToKebab(schema.Name)),
		UniqueFields:          uniqueFields,
		Fields:                fields,
		ModelFields:           modelFields,
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
