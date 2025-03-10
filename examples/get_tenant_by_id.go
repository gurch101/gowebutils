package main

import (
	"context"
	"net/http"

	"github.com/gurch101/gowebutils/pkg/app"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/parser"
)

type GetTenantResponse struct {
	ID           int64      `json:"id"`
	TenantName   string     `json:"tenantName"`
	ContactEmail string     `json:"contactEmail"`
	Plan         TenantPlan `json:"plan"`
	IsActive     bool       `json:"isActive"`
}

type GetTenantController struct {
	app *app.App
}

func NewGetTenantController(app *app.App) *GetTenantController {
	return &GetTenantController{app: app}
}

func (tc *GetTenantController) GetTenantHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parser.ReadIDPathParam(r)

	if err != nil {
		httputils.NotFoundResponse(w, r)
		return
	}

	tenant, err := GetTenantByID(r.Context(), tc.app.DB, id)

	if err != nil {
		httputils.HandleErrorResponse(w, r, err)
		return
	}

	err = httputils.WriteJSON(w, http.StatusOK, &GetTenantResponse{
		ID:           tenant.ID,
		TenantName:   tenant.TenantName,
		ContactEmail: tenant.ContactEmail,
		Plan:         tenant.Plan,
		IsActive:     tenant.IsActive,
	}, nil)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

func GetTenantByID(ctx context.Context, db dbutils.DB, tenantID int64) (*tenantModel, error) {
	var tenant tenantModel

	err := dbutils.GetByID(ctx, db, tenantResourceKey, tenantID, map[string]any{
		tenantIDDBFieldName:     &tenant.ID,
		tenantNameDBFieldName:   &tenant.TenantName,
		contactEmailDBFieldName: &tenant.ContactEmail,
		planDBFieldName:         &tenant.Plan,
		isActiveDBFieldName:     &tenant.IsActive,
		createdAtDBFieldName:    &tenant.CreatedAt,
		versionDBFieldName:      &tenant.Version,
	})
	if err != nil {
		return nil, dbutils.WrapDBError(err)
	}
	return &tenant, nil
}
