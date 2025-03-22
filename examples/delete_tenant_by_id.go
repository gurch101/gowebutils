package main

import (
	"context"
	"net/http"

	"github.com/gurch101/gowebutils/pkg/app"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/parser"
)

type DeleteTenantController struct {
	app *app.App
}

func NewDeleteTenantController(app *app.App) *DeleteTenantController {
	return &DeleteTenantController{app: app}
}

func (tc *DeleteTenantController) DeleteTenantHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parser.ReadIDPathParam(r)

	if err != nil {
		httputils.NotFoundResponse(w, r)

		return
	}

	err = DeleteTenantByID(r.Context(), tc.app.DB(), id)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)

		return
	}

	err = httputils.WriteJSON(w, http.StatusOK, envelope{"message": "Tenant successfully deleted"}, nil)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

func DeleteTenantByID(ctx context.Context, db dbutils.DB, tenantId int64) error {
	return dbutils.DeleteByID(ctx, db, tenantResourceKey, tenantId)
}
