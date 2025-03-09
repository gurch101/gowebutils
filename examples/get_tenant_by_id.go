package main

import (
	"context"
	"net/http"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/parser"
)

func (tc *TenantController) getTenantByIdHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parser.ReadIDPathParam(r)

	if err != nil {
		httputils.NotFoundResponse(w, r)
		return
	}

	tenant, err := getTenantById(tc.appserver.DB, id)

	if err != nil {
		httputils.HandleErrorResponse(w, r, err)
		return
	}

	err = httputils.WriteJSON(w, http.StatusOK, &GetTenantResponse{ID: tenant.ID, TenantName: tenant.TenantName, ContactEmail: tenant.ContactEmail, Plan: tenant.Plan, IsActive: tenant.IsActive}, nil)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

func getTenantById(db dbutils.DB, tenantId int64) (*tenantModel, error) {
	var tenant tenantModel

	err := dbutils.GetByID(context.Background(), db, tenantResourceKey, tenantId, map[string]any{
		tenantIdDbFieldName:     &tenant.ID,
		tenantNameDbFieldName:   &tenant.TenantName,
		contactEmailDbFieldName: &tenant.ContactEmail,
		planDbFieldName:         &tenant.Plan,
		isActiveDbFieldName:     &tenant.IsActive,
		createdAtDbFieldName:    &tenant.CreatedAt,
		versionDbFieldName:      &tenant.Version,
	})
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}
