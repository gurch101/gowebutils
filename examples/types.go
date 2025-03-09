package main

import (
	"time"

	"github.com/gurch101/gowebutils/pkg/parser"
	"github.com/gurch101/gowebutils/pkg/validation"
)

const (
	tenantNameRequestKey   = "tenantName"
	planRequestKey         = "plan"
	contactEmailRequestKey = "contactEmail"
	tenantResourceKey      = "tenants"
)

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

var ErrTenantAlreadyRegistered = validation.Error{
	Field:   tenantNameRequestKey,
	Message: "This tenant is already registered",
}

type CreateTenantRequest struct {
	TenantName   string     `json:"tenantName"`
	ContactEmail string     `json:"contactEmail"`
	Plan         TenantPlan `json:"plan"`
}

type GetTenantResponse struct {
	ID           int64      `json:"id"`
	TenantName   string     `json:"tenantName"`
	ContactEmail string     `json:"contactEmail"`
	Plan         TenantPlan `json:"plan"`
	IsActive     bool       `json:"isActive"`
}

type UpdateTenantRequest struct {
	TenantName   *string     `json:"tenantName"`
	ContactEmail *string     `json:"contactEmail"`
	Plan         *TenantPlan `json:"plan"`
	IsActive     *bool       `json:"isActive"`
}

type SearchTenantsRequest struct {
	TenantName   *string
	Plan         *string
	IsActive     *bool
	ContactEmail *string
	parser.Filters
}

type SearchTenantResponse struct {
	ID           int64      `json:"id"`
	TenantName   string     `json:"tenantName"`
	ContactEmail string     `json:"contactEmail"`
	Plan         TenantPlan `json:"plan"`
	IsActive     bool       `json:"isActive"`
	CreatedAt    time.Time  `json:"createdAt"`
}
