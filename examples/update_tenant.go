package main

import (
	"context"
	"net/http"

	"github.com/gurch101/gowebutils/pkg/app"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/parser"
	"github.com/gurch101/gowebutils/pkg/validation"
)

type UpdateTenantController struct {
	app *app.App
}

func NewUpdateTenantController(app *app.App) *UpdateTenantController {
	return &UpdateTenantController{app: app}
}

type UpdateTenantRequest struct {
	TenantName   *string     `json:"tenantName"`
	ContactEmail *string     `json:"contactEmail"`
	Plan         *TenantPlan `json:"plan"`
	IsActive     *bool       `json:"isActive"`
}

func (tc *UpdateTenantController) UpdateTenantHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parser.ParseIDPathParam(r)

	if err != nil {
		httputils.NotFoundResponse(w, r)

		return
	}

	tenant, err := GetTenantByID(r.Context(), tc.app.DB(), id)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)

		return
	}

	updateTenantRequest, err := httputils.ReadJSON[UpdateTenantRequest](w, r)
	if err != nil {
		httputils.UnprocessableEntityResponse(w, r, err)

		return
	}

	tenant.TenantName = validation.Coalesce(updateTenantRequest.TenantName, tenant.TenantName)
	tenant.ContactEmail = validation.Coalesce(updateTenantRequest.ContactEmail, tenant.ContactEmail)
	tenant.Plan = validation.Coalesce(updateTenantRequest.Plan, tenant.Plan)
	tenant.IsActive = validation.Coalesce(updateTenantRequest.IsActive, tenant.IsActive)

	v := validation.NewValidator()
	v.Required(tenant.TenantName, tenantNameRequestKey, "Tenant Name is required")
	v.Email(tenant.ContactEmail, contactEmailRequestKey, "Contact Email is required")
	v.Check(IsValidTenantPlan(tenant.Plan), planRequestKey, "Invalid plan")

	if v.HasErrors() {
		httputils.FailedValidationResponse(w, r, v.Errors)

		return
	}

	err = UpdateTenant(r.Context(), tc.app.DB(), tenant)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)

		return
	}

	err = httputils.WriteJSON(
		w,
		http.StatusOK,
		&GetTenantResponse{
			ID:           tenant.ID,
			TenantName:   tenant.TenantName,
			ContactEmail: tenant.ContactEmail,
			Plan:         tenant.Plan,
			IsActive:     tenant.IsActive,
		},
		nil)

	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

func UpdateTenant(ctx context.Context, db dbutils.DB, tenant *tenantModel) error {
	return dbutils.UpdateByID(ctx, db, tenantResourceKey, tenant.ID, tenant.Version, map[string]any{
		tenantNameDBFieldName:   tenant.TenantName,
		contactEmailDBFieldName: tenant.ContactEmail,
		planDBFieldName:         tenant.Plan,
		isActiveDBFieldName:     tenant.IsActive,
	})
}
