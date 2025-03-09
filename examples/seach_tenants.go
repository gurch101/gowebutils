package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/parser"
	"github.com/gurch101/gowebutils/pkg/validation"
)

func (tc *TenantController) searchTenantsHandler(w http.ResponseWriter, r *http.Request) {
	v := validation.NewValidator()
	queryString := r.URL.Query()
	searchTenantsRequest := &SearchTenantsRequest{
		TenantName:   parser.ParseQSString(queryString, tenantNameRequestKey, nil),
		Plan:         parser.ParseQSString(queryString, planRequestKey, nil),
		IsActive:     parser.ParseQSBool(queryString, "isActive", nil),
		ContactEmail: parser.ParseQSString(queryString, contactEmailRequestKey, nil),
	}

	searchTenantsRequest.ParseQSMetadata(queryString, v, []string{"id", tenantNameRequestKey, planRequestKey, contactEmailRequestKey, fmt.Sprintf("-%s", tenantNameRequestKey), fmt.Sprintf("-%s", planRequestKey), fmt.Sprintf("-%s", contactEmailRequestKey)})
	if v.HasErrors() {
		httputils.FailedValidationResponse(w, r, v.Errors)
		return
	}

	tenants, pagination, err := SearchTenants(tc.appserver.DB, searchTenantsRequest)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)
		return
	}
	err = httputils.WriteJSON(w, http.StatusOK, envelope{"metadata": pagination, tenantResourceKey: tenants}, nil)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

func SearchTenants(db *sql.DB, searchTenantsRequest *SearchTenantsRequest) ([]*SearchTenantResponse, parser.PaginationMetadata, error) {
	tenants, pagination, err := FindTenants(db, searchTenantsRequest)
	if err != nil {
		return nil, pagination, err
	}

	tenantResponses := make([]*SearchTenantResponse, 0)

	for _, tenant := range tenants {
		tenantResponse := &SearchTenantResponse{
			ID:           tenant.ID,
			TenantName:   tenant.TenantName,
			ContactEmail: tenant.ContactEmail,
			Plan:         tenant.Plan,
			IsActive:     tenant.IsActive,
			CreatedAt:    tenant.CreatedAt,
		}
		tenantResponses = append(tenantResponses, tenantResponse)
	}
	return tenantResponses, pagination, nil
}
