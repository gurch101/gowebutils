package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/validation"
)

func (tc *TenantController) createTenantHandler(w http.ResponseWriter, r *http.Request) {
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

	tenantID, err := CreateTenant(tc.appserver.DB, &createTenantRequest)
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

func CreateTenant(db dbutils.DB, createTenantRequest *CreateTenantRequest) (*int64, error) {
	tenantModel := NewTenantModel(createTenantRequest.TenantName, createTenantRequest.ContactEmail, createTenantRequest.Plan)

	id, err := insertTenant(db, tenantModel)

	if err != nil {
		if errors.Is(err, dbutils.ErrUniqueConstraint) && strings.Contains(err.Error(), tenantNameDbFieldName) {
			return nil, ErrTenantAlreadyRegistered
		}

		return nil, err
	}
	return id, nil
}

func insertTenant(db dbutils.DB, tenant *tenantModel) (*int64, error) {
	return dbutils.Insert(context.Background(), db, tenantResourceKey, map[string]any{
		tenantNameDbFieldName:   tenant.TenantName,
		contactEmailDbFieldName: tenant.ContactEmail,
		planDbFieldName:         tenant.Plan,
		isActiveDbFieldName:     tenant.IsActive,
	})
}
