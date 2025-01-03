package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"gurch101.github.io/go-web/pkg/dbutils"
	"gurch101.github.io/go-web/pkg/httputils"
	"gurch101.github.io/go-web/pkg/parser"
	"gurch101.github.io/go-web/pkg/validation"
)

type envelope map[string]interface{}

type Controller interface {
	GetMux() *http.ServeMux
}

type TenantController struct {
	DB *sql.DB
}

func NewTenantController(db *sql.DB) *TenantController {
	return &TenantController{DB: db}
}

func (c *TenantController) GetMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tenants", c.CreateTenantHandler)
	mux.HandleFunc("GET /tenants/{id}", c.GetTenantHandler)
	mux.HandleFunc("GET /tenants", c.SearchTenantsHandler)
	mux.HandleFunc("PATCH /tenants/{id}", c.UpdateTenantHandler)
	mux.HandleFunc("DELETE /tenants/{id}", c.DeleteTenantHandler)
	return mux
}

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

type CreateTenantRequest struct {
	TenantName   string     `json:"tenantName"`
	ContactEmail string     `json:"contactEmail"`
	Plan         TenantPlan `json:"plan"`
}

func (tc *TenantController) CreateTenantHandler(w http.ResponseWriter, r *http.Request) {
	createTenantRequest := &CreateTenantRequest{}
	err := parser.ReadJSON(w, r, createTenantRequest)
	if err != nil {
		httputils.UnprocessableEntityResponse(w, r, err)
		return
	}

	v := validation.NewValidator()
	v.Required(createTenantRequest.TenantName, "tenantName", "Tenant Name is required")
	v.Email(createTenantRequest.ContactEmail, "contactEmail", "Contact Email is required")
	v.Check(IsValidTenantPlan(createTenantRequest.Plan), "plan", "Invalid plan")

	if v.HasErrors() {
		httputils.FailedValidationResponse(w, r, v.Errors)
		return
	}

	tenantId, err := CreateTenant(tc.DB, createTenantRequest)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/tenants/%d", tenantId))
	err = parser.WriteJSON(w, http.StatusCreated, envelope{"id": tenantId}, headers)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

type GetTenantResponse struct {
	ID           int64      `json:"id"`
	TenantName   string     `json:"tenantName"`
	ContactEmail string     `json:"contactEmail"`
	Plan         TenantPlan `json:"plan"`
	IsActive     bool       `json:"isActive"`
}

func (tc *TenantController) GetTenantHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parser.ReadIDPathParam(r)

	if err != nil {
		httputils.NotFoundResponse(w, r)
		return
	}

	tenant, err := GetTenantById(tc.DB, id)

	if err != nil {
		httputils.HandleErrorResponse(w, r, err)
		return
	}

	err = parser.WriteJSON(w, http.StatusOK, &GetTenantResponse{ID: tenant.ID, TenantName: tenant.TenantName, ContactEmail: tenant.ContactEmail, Plan: tenant.Plan, IsActive: tenant.IsActive}, nil)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

type UpdateTenantRequest struct {
	TenantName   *string     `json:"tenantName"`
	ContactEmail *string     `json:"contactEmail"`
	Plan         *TenantPlan `json:"plan"`
	IsActive     *bool       `json:"isActive"`
}

func (tc *TenantController) UpdateTenantHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parser.ReadIDPathParam(r)

	if err != nil {
		httputils.NotFoundResponse(w, r)
		return
	}

	tenant, err := GetTenantById(tc.DB, id)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)
		return
	}

	updateTenantRequest := &UpdateTenantRequest{}
	err = parser.ReadJSON(w, r, updateTenantRequest)
	if err != nil {
		httputils.UnprocessableEntityResponse(w, r, err)
		return
	}
	tenant.TenantName = validation.Coalesce(updateTenantRequest.TenantName, tenant.TenantName)
	tenant.ContactEmail = validation.Coalesce(updateTenantRequest.ContactEmail, tenant.ContactEmail)
	tenant.Plan = validation.Coalesce(updateTenantRequest.Plan, tenant.Plan)
	tenant.IsActive = validation.Coalesce(updateTenantRequest.IsActive, tenant.IsActive)

	v := validation.NewValidator()
	v.Required(tenant.TenantName, "tenantName", "Tenant Name is required")
	v.Email(tenant.ContactEmail, "contactEmail", "Contact Email is required")
	v.Check(IsValidTenantPlan(tenant.Plan), "plan", "Invalid plan")

	if v.HasErrors() {
		httputils.FailedValidationResponse(w, r, v.Errors)
		return
	}

	err = UpdateTenant(tc.DB, tenant)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)
		return
	}

	err = parser.WriteJSON(w, http.StatusOK, &GetTenantResponse{ID: tenant.ID, TenantName: tenant.TenantName, ContactEmail: tenant.ContactEmail, Plan: tenant.Plan, IsActive: tenant.IsActive}, nil)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

func (tc *TenantController) DeleteTenantHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parser.ReadIDPathParam(r)

	if err != nil {
		httputils.NotFoundResponse(w, r)
		return
	}

	err = DeleteTenantById(tc.DB, id)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)
		return
	}

	err = parser.WriteJSON(w, http.StatusOK, envelope{"message": "Tenant successfully deleted"}, nil)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

type SearchTenantsRequest struct {
	TenantName   *string
	Plan         *string
	IsActive     *bool
	ContactEmail *string
	parser.Filters
}

func (tc *TenantController) SearchTenantsHandler(w http.ResponseWriter, r *http.Request) {
	v := validation.NewValidator()
	qs := r.URL.Query()
	searchTenantsRequest := &SearchTenantsRequest{}
	searchTenantsRequest.TenantName = parser.ParseString(qs, "tenantName", nil)
	searchTenantsRequest.Plan = parser.ParseString(qs, "plan", nil)
	searchTenantsRequest.IsActive = parser.ParseBool(qs, "isActive", nil)
	searchTenantsRequest.ContactEmail = parser.ParseString(qs, "contactEmail", nil)
	searchTenantsRequest.ParseFilters(qs, v, []string{"id", "tenantName", "plan", "contactEmail", "-tenantName", "-plan", "-contactEmail"})
	if v.HasErrors() {
		httputils.FailedValidationResponse(w, r, v.Errors)
		return
	}

	tenants, pagination, err := SearchTenants(tc.DB, searchTenantsRequest)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)
		return
	}
	err = parser.WriteJSON(w, http.StatusOK, envelope{"metadata": pagination, "tenants": tenants}, nil)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

// service layer
func CreateTenant(db *sql.DB, createTenantRequest *CreateTenantRequest) (*int64, error) {
	tenantModel := NewTenantModel(createTenantRequest.TenantName, createTenantRequest.ContactEmail, createTenantRequest.Plan)

	id, err := InsertTenant(db, tenantModel)

	if err != nil {
		if errors.As(err, &dbutils.ConstraintError{}) {
			if err.(dbutils.ConstraintError).DetailContains("tenant_name") {
				return nil, validation.ValidationError{Field: "tenantName", Message: "This tenant is already registered"}
			}
		}
		return nil, err
	}
	return id, nil
}

type SearchTenantResponse struct {
	ID           int64      `json:"id"`
	TenantName   string     `json:"tenantName"`
	ContactEmail string     `json:"contactEmail"`
	Plan         TenantPlan `json:"plan"`
	IsActive     bool       `json:"isActive"`
	CreatedAt    time.Time  `json:"createdAt"`
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

// repository layer
type tenantModel struct {
	ID           int64
	TenantName   string
	ContactEmail string
	Plan         TenantPlan
	IsActive     bool
	CreatedAt    time.Time
	Version      int32
}

func NewTenantModel(name, email string, plan TenantPlan) *tenantModel {
	return &tenantModel{
		TenantName:   name,
		ContactEmail: email,
		Plan:         plan,
		IsActive:     true,
	}
}

func InsertTenant(db *sql.DB, tenant *tenantModel) (*int64, error) {
	return dbutils.Insert(db, "tenants", map[string]any{
		"tenant_name":   tenant.TenantName,
		"contact_email": tenant.ContactEmail,
		"plan":          tenant.Plan,
		"is_active":     tenant.IsActive,
	})
}

func GetTenantById(db *sql.DB, tenantId int64) (*tenantModel, error) {
	var tenant tenantModel

	err := dbutils.GetById(db, "tenants", tenantId, map[string]any{
		"id":            &tenant.ID,
		"tenant_name":   &tenant.TenantName,
		"contact_email": &tenant.ContactEmail,
		"plan":          &tenant.Plan,
		"is_active":     &tenant.IsActive,
		"created_at":    &tenant.CreatedAt,
		"version":       &tenant.Version,
	})
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func DeleteTenantById(db *sql.DB, tenantId int64) error {
	return dbutils.DeleteById(db, "tenants", tenantId)
}

func UpdateTenant(db *sql.DB, tenant *tenantModel) error {
	return dbutils.UpdateById(db, "tenants", tenant.ID, tenant.Version, map[string]any{
		"tenant_name":   tenant.TenantName,
		"contact_email": tenant.ContactEmail,
		"plan":          tenant.Plan,
		"is_active":     tenant.IsActive,
	})
}

func FindTenants(db *sql.DB, searchTenantsRequest *SearchTenantsRequest) ([]tenantModel, parser.PaginationMetadata, error) {
	var tenants []tenantModel
	var totalRecords int
	err := dbutils.NewQueryBuilder(db).
		Select("count(*) over()", "id", "tenant_name", "contact_email", "plan", "is_active", "created_at", "version").
		From("tenants").
		WhereLike("tenant_name", dbutils.OpContains, searchTenantsRequest.TenantName).
		AndWhere("plan = ?", searchTenantsRequest.Plan).
		AndWhere("is_active = ?", searchTenantsRequest.IsActive).
		AndWhereLike("contact_email", dbutils.OpContains, searchTenantsRequest.ContactEmail).
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

func main() {
	logger := httputils.InitializeSlog(os.Getenv("LOG_LEVEL"))

	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?_foreign_keys=1&_journal=WAL", os.Getenv("DB_FILEPATH")))
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}

	defer db.Close()

	slog.Info("Starting server on :8080")

	tenantController := NewTenantController(db)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      httputils.LoggingMiddleware(httputils.RecoveryMiddleware(tenantController.GetMux())),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     httputils.NewSlogErrorWriter(logger),
	}
	err = server.ListenAndServe()
	if err != nil {
		slog.Error(err.Error())
	}
}
