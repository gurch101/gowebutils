package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/parser"
)

type tenantModel struct {
	ID           int64
	TenantName   string
	ContactEmail string
	Plan         TenantPlan
	IsActive     bool
	CreatedAt    time.Time
	Version      int32
}

const (
	tenantIdDbFieldName     = "id"
	tenantNameDbFieldName   = "tenant_name"
	contactEmailDbFieldName = "contact_email"
	planDbFieldName         = "plan"
	isActiveDbFieldName     = "is_active"
	createdAtDbFieldName    = "created_at"
	versionDbFieldName      = "version"
)

func NewTenantModel(name, email string, plan TenantPlan) *tenantModel {
	return &tenantModel{
		TenantName:   name,
		ContactEmail: email,
		Plan:         plan,
		IsActive:     true,
	}
}

func InsertTenant(db *sql.DB, tenant *tenantModel) (*int64, error) {
	return dbutils.Insert(context.Background(), db, tenantResourceKey, map[string]any{
		tenantNameDbFieldName:   tenant.TenantName,
		contactEmailDbFieldName: tenant.ContactEmail,
		planDbFieldName:         tenant.Plan,
		isActiveDbFieldName:     tenant.IsActive,
	})
}

func GetTenantById(db *sql.DB, tenantId int64) (*tenantModel, error) {
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

func DeleteTenantById(db *sql.DB, tenantId int64) error {
	return dbutils.DeleteByID(context.Background(), db, tenantResourceKey, tenantId)
}

func UpdateTenant(db *sql.DB, tenant *tenantModel) error {
	return dbutils.UpdateByID(db, tenantResourceKey, tenant.ID, tenant.Version, map[string]any{
		tenantNameDbFieldName:   tenant.TenantName,
		contactEmailDbFieldName: tenant.ContactEmail,
		planDbFieldName:         tenant.Plan,
		isActiveDbFieldName:     tenant.IsActive,
	})
}

func FindTenants(db *sql.DB, searchTenantsRequest *SearchTenantsRequest) ([]tenantModel, parser.PaginationMetadata, error) {
	var tenants []tenantModel
	var totalRecords int
	err := dbutils.NewQueryBuilder(db).
		Select("count(*) over()", tenantIdDbFieldName, tenantNameDbFieldName, contactEmailDbFieldName, planDbFieldName, isActiveDbFieldName, createdAtDbFieldName, versionDbFieldName).
		From(tenantResourceKey).
		WhereLike(tenantNameDbFieldName, dbutils.OpContains, searchTenantsRequest.TenantName).
		AndWhere(fmt.Sprintf("%s = ?", planDbFieldName), searchTenantsRequest.Plan).
		AndWhere(fmt.Sprintf("%s = ?", isActiveDbFieldName), searchTenantsRequest.IsActive).
		AndWhereLike(contactEmailDbFieldName, dbutils.OpContains, searchTenantsRequest.ContactEmail).
		OrderBy(searchTenantsRequest.Sort).
		Page(searchTenantsRequest.Page, searchTenantsRequest.PageSize).
		Execute(func(rows *sql.Rows) error {
			var tenant tenantModel
			err := rows.Scan(&totalRecords, &tenant.ID, &tenant.TenantName, &tenant.ContactEmail, &tenant.Plan, &tenant.IsActive, &tenant.CreatedAt, &tenant.Version)
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
