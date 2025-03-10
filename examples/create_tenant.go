package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gurch101/gowebutils/pkg/app"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/validation"
)

type CreateTenantController struct {
	app *app.App
}

var ErrTenantAlreadyRegistered = validation.Error{
	Field:   tenantNameRequestKey,
	Message: "This tenant is already registered",
}

type CreateTenantRequest struct {
	TenantName   string     `json:"tenantName"`
	ContactEmail string     `json:"contactEmail"`
	Plan         TenantPlan `json:"plan"`
}

func NewCreateTenantController(app *app.App) *CreateTenantController {
	return &CreateTenantController{app: app}
}

func (c *CreateTenantController) CreateTenantHandler(w http.ResponseWriter, r *http.Request) {
	createTenantRequest, err := httputils.ReadJSON[CreateTenantRequest](w, r)
	if err != nil {
		httputils.UnprocessableEntityResponse(w, r, err)

		return
	}

	v := validation.NewValidator()
	v.Required(createTenantRequest.TenantName, tenantNameRequestKey, "Tenant Name is required")
	v.Email(createTenantRequest.ContactEmail, contactEmailRequestKey, "Contact Email is required")
	v.Check(IsValidTenantPlan(createTenantRequest.Plan), planRequestKey, "Invalid plan")

	if v.HasErrors() {
		httputils.FailedValidationResponse(w, r, v.Errors)

		return
	}

	tenantID, err := CreateTenant(r.Context(), c.app.DB, &createTenantRequest)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)

		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/tenants/%d", tenantID))

	err = httputils.WriteJSON(w, http.StatusCreated, envelope{"id": tenantID}, headers)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

func CreateTenant(ctx context.Context, db dbutils.DB, createTenantRequest *CreateTenantRequest) (*int64, error) {
	tenantModel := newTenantModel(createTenantRequest.TenantName, createTenantRequest.ContactEmail, createTenantRequest.Plan)

	id, err := insertTenant(ctx, db, tenantModel)

	if err != nil {
		if errors.Is(err, dbutils.ErrUniqueConstraint) && strings.Contains(err.Error(), tenantNameDBFieldName) {
			return nil, ErrTenantAlreadyRegistered
		}

		return nil, err
	}
	return id, nil
}

func insertTenant(ctx context.Context, db dbutils.DB, tenant *tenantModel) (*int64, error) {
	return dbutils.Insert(ctx, db, tenantResourceKey, map[string]any{
		tenantNameDBFieldName:   tenant.TenantName,
		contactEmailDBFieldName: tenant.ContactEmail,
		planDBFieldName:         tenant.Plan,
		isActiveDBFieldName:     tenant.IsActive,
	})
}
