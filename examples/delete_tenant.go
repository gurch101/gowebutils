package main

import (
	"context"
	"net/http"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/parser"
)

func (tc *TenantController) deleteTenantHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parser.ReadIDPathParam(r)

	if err != nil {
		httputils.NotFoundResponse(w, r)

		return
	}

	err = DeleteTenantById(tc.appserver.DB, id)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)

		return
	}

	err = httputils.WriteJSON(w, http.StatusOK, envelope{"message": "Tenant successfully deleted"}, nil)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

func DeleteTenantById(db dbutils.DB, tenantId int64) error {
	return dbutils.DeleteByID(context.Background(), db, tenantResourceKey, tenantId)
}
