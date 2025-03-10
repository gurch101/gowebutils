package main

import "time"

type TenantPlan string

const (
	Free TenantPlan = "free"
	Paid TenantPlan = "paid"
)

func IsValidTenantPlan(plan TenantPlan) bool {
	switch plan {
	case Free, Paid:
		return true
	}

	return false
}

const (
	tenantNameRequestKey   = "tenantName"
	planRequestKey         = "plan"
	contactEmailRequestKey = "contactEmail"
	tenantResourceKey      = "tenants"
)

const (
	tenantIDDBFieldName     = "id"
	tenantNameDBFieldName   = "tenant_name"
	contactEmailDBFieldName = "contact_email"
	planDBFieldName         = "plan"
	isActiveDBFieldName     = "is_active"
	createdAtDBFieldName    = "created_at"
	versionDBFieldName      = "version"
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

func newTenantModel(name, email string, plan TenantPlan) *tenantModel {
	return &tenantModel{
		TenantName:   name,
		ContactEmail: email,
		Plan:         plan,
		IsActive:     true,
	}
}
