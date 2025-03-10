package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gurch101/gowebutils/pkg/app"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/parser"
	"github.com/gurch101/gowebutils/pkg/validation"
)

type SearchTenantController struct {
	app *app.App
}

func NewSearchTenantController(app *app.App) *SearchTenantController {
	return &SearchTenantController{app: app}
}

type SearchTenantsRequest struct {
	TenantName   *string
	Plan         *string
	IsActive     *bool
	ContactEmail *string
	parser.Filters
}

func (tc *SearchTenantController) SearchTenantsHandler(w http.ResponseWriter, r *http.Request) {
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

	tenants, pagination, err := SearchTenants(tc.app.DB, searchTenantsRequest)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)
		return
	}
	err = httputils.WriteJSON(w, http.StatusOK, envelope{"metadata": pagination, tenantResourceKey: tenants}, nil)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

type SearchTenantResponse struct {
	ID           int64      `json:"id"`
	TenantName   string     `json:"tenantName"`
	ContactEmail string     `json:"contactEmail"`
	Plan         TenantPlan `json:"plan"`
	IsActive     bool       `json:"isActive"`
	CreatedAt    time.Time  `json:"createdAt"`
}

func SearchTenants(db dbutils.DB, searchTenantsRequest *SearchTenantsRequest) ([]*SearchTenantResponse, parser.PaginationMetadata, error) {
	tenants, pagination, err := findTenants(db, searchTenantsRequest)
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

func findTenants(
	db dbutils.DB,
	searchTenantsRequest *SearchTenantsRequest) ([]tenantModel, parser.PaginationMetadata, error) {
	var tenants []tenantModel
	var totalRecords int
	err := dbutils.NewQueryBuilder(db).
		Select(
			"count(*) over()",
			tenantIDDBFieldName,
			tenantNameDBFieldName,
			contactEmailDBFieldName,
			planDBFieldName,
			isActiveDBFieldName,
			createdAtDBFieldName,
			versionDBFieldName,
		).
		From(tenantResourceKey).
		WhereLike(tenantNameDBFieldName, dbutils.OpContains, searchTenantsRequest.TenantName).
		AndWhere(planDBFieldName+" = ?", searchTenantsRequest.Plan).
		AndWhere(isActiveDBFieldName+" = ?", searchTenantsRequest.IsActive).
		AndWhereLike(contactEmailDBFieldName, dbutils.OpContains, searchTenantsRequest.ContactEmail).
		OrderBy(searchTenantsRequest.Sort).
		Page(searchTenantsRequest.Page, searchTenantsRequest.PageSize).
		Exec(func(rows *sql.Rows) error {
			var tenant tenantModel
			err := rows.Scan(
				&totalRecords,
				&tenant.ID,
				&tenant.TenantName,
				&tenant.ContactEmail,
				&tenant.Plan,
				&tenant.IsActive,
				&tenant.CreatedAt,
				&tenant.Version,
			)

			if err != nil {
				return err
			}

			tenants = append(tenants, tenant)
			return nil
		})

	if err != nil {
		return nil, parser.PaginationMetadata{}, dbutils.WrapDBError(err)
	}

	metadata := parser.ParsePaginationMetadata(totalRecords, searchTenantsRequest.Page, searchTenantsRequest.PageSize)
	return tenants, metadata, nil
}
