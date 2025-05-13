package generator

import (
	"fmt"
	"strings"

	"github.com/gurch101/gowebutils/pkg/stringutils"
)

const deleteHandlerTemplate = `package {{.PackageName}}

import (
	"context"
	"net/http"

	"github.com/gurch101/gowebutils/pkg/app"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/parser"
)

type Delete{{.SingularTitleCaseName}}Controller struct {
	app *app.App
}

func NewDelete{{.SingularTitleCaseName}}Controller(app *app.App) *Delete{{.SingularTitleCaseName}}Controller {
	return &Delete{{.SingularTitleCaseName}}Controller{app: app}
}

type Delete{{.SingularTitleCaseName}}Response struct {
	Message string ` + "`" + `json:"message"` + "`" + `
}

// Delete{{.SingularTitleCaseName}} godoc
//
//	@Summary		Delete a {{.HumanName}}
//	@Description	Delete by {{.HumanName}} ID
//	@Tags			{{.HumanName}}s
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"{{.SingularCamelCaseName}} ID"	Format(int64)
//	@Success		200	{object}	Delete{{.SingularTitleCaseName}}Response
//	@Failure		400,404,422,500	{object}	httputils.ErrorResponse
//	@Router			/{{.KebabCaseTableName}}/{id} [delete]
func (tc *Delete{{.SingularTitleCaseName}}Controller) Delete{{.SingularTitleCaseName}}Handler(w http.ResponseWriter, r *http.Request) {
	id, err := parser.ParseIDPathParam(r)

	if err != nil {
		httputils.NotFoundResponse(w, r)

		return
	}

	err = Delete{{.SingularTitleCaseName}}ByID(r.Context(), tc.app.DB(), id)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)

		return
	}

	err = httputils.WriteJSON(w, http.StatusOK, Delete{{.SingularTitleCaseName}}Response{Message: "{{.SingularTitleCaseName}} successfully deleted"}, nil)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

func Delete{{.SingularTitleCaseName}}ByID(ctx context.Context, db dbutils.DB, id int64) error {
	return dbutils.DeleteByID(ctx, db, "{{.Name}}", id)
}
`

const deleteHandlerTestTemplate = `package {{.PackageName}}_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"{{.ModuleName}}/internal/{{.PackageName}}"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestDelete{{.SingularTitleCaseName}}(t *testing.T) {
	t.Parallel()

	t.Run("successful delete", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		defer app.Close()

		deleteController := {{.PackageName}}.NewDelete{{.SingularTitleCaseName}}Controller(app.App)

		app.TestRouter.Delete("/{{.KebabCaseTableName}}/{id}", deleteController.Delete{{.SingularTitleCaseName}}Handler)

		ID, _ := {{.PackageName}}.CreateTest{{.SingularTitleCaseName}}(t, app.DB())

		deleteURL := fmt.Sprintf("/{{.KebabCaseTableName}}/%d", ID)
		deleteReq := testutils.CreateDeleteRequest(deleteURL)
		deleteRr := app.MakeRequest(deleteReq)

		if deleteRr.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, deleteRr.Code)
		}

		var deleteResponse {{.PackageName}}.Delete{{.SingularTitleCaseName}}Response
		err = json.Unmarshal(deleteRr.Body.Bytes(), &deleteResponse)
		if err != nil {
			t.Fatal(err)
		}

		if deleteResponse.Message != "{{.SingularTitleCaseName}} successfully deleted" {
			t.Errorf("expected message to be '{{.SingularTitleCaseName}} successfully deleted', got '%s'", deleteResponse.Message)
		}

		var count int
		err = app.DB().QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM {{.Name}} WHERE id = $1", ID).Scan(&count)
		if err != nil {
			t.Fatal(err)
		}

		if count != 0 {
			t.Errorf("expected record to be deleted, but it still exists in the database")
		}
	})

	t.Run("delete non-existent record", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		defer app.Close()

		controller := {{.PackageName}}.NewDelete{{.SingularTitleCaseName}}Controller(app.App)
		app.TestRouter.Delete("/{{.KebabCaseTableName}}/{id}", controller.Delete{{.SingularTitleCaseName}}Handler)

		// Use a non-existent ID
		nonExistentID := int64(99999)
		deleteURL := fmt.Sprintf("/{{.KebabCaseTableName}}/%d", nonExistentID)
		deleteReq := testutils.CreateDeleteRequest(deleteURL)
		rr := app.MakeRequest(req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d", http.StatusNotFound, rr.Code)
		}
	})

	t.Run("invalid ID format", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		defer app.Close()

		controller := {{.PackageName}}.NewDelete{{.SingularTitleCaseName}}Controller(app.App)
		app.TestRouter.Delete("/{{.KebabCaseTableName}}/{id}", controller.Delete{{.SingularTitleCaseName}}Handler)

		// Use an invalid ID format
		deleteURL := "/{{.KebabCaseTableName}}/invalid-id"
		deleteReq := testutils.CreateDeleteRequest(deleteURL)
		rr := app.MakeRequest(req)

		// Should return 404 Not Found for invalid ID format
		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d", http.StatusNotFound, rr.Code)
		}
	})
}
`

func newDeleteHandlerTemplateData(moduleName string, schema Table) deleteHandlerTemplateData {
	var fields []RequestField

	for _, field := range schema.Fields {
		sanitizedName := field.Name
		if strings.HasSuffix(field.Name, "id") {
			sanitizedName = strings.TrimSuffix(field.Name, "id") + "ID"
		}

		if IsRequestField(field) {
			required := hasBlankConstraint(field.Constraints)
			email := isEmail(sanitizedName)

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
	}

	return deleteHandlerTemplateData{
		PackageName:           schema.Name,
		Name:                  schema.Name,
		ModuleName:            moduleName,
		HumanName:             stringutils.SnakeToHuman(strings.TrimSuffix(schema.Name, "s")),
		TitleCaseTableName:    stringutils.SnakeToTitle(schema.Name),
		KebabCaseTableName:    stringutils.SnakeToKebab(schema.Name),
		SingularTitleCaseName: stringutils.SnakeToTitle(strings.TrimSuffix(schema.Name, "s")),
		SingularCamelCaseName: stringutils.SnakeToCamel(strings.TrimSuffix(schema.Name, "s")),
		CreateFields:          fields,
	}
}

func RenderDeleteTemplate(moduleName string, schema Table) ([]byte, []byte, error) {
	data := newDeleteHandlerTemplateData(moduleName, schema)

	tmpl, err := renderTemplateFile(deleteHandlerTemplate, data)
	if err != nil {
		return nil, nil, fmt.Errorf("error rendering create template: %w", err)
	}

	testTmpl, err := renderTemplateFile(deleteHandlerTestTemplate, data)
	if err != nil {
		return nil, nil, fmt.Errorf("error rendering create test template: %w", err)
	}

	return tmpl, testTmpl, nil
}
