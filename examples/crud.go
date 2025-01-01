package main

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"gurch101.github.io/go-web/pkg/dbutils"
	"gurch101.github.io/go-web/pkg/middleware"
	"gurch101.github.io/go-web/pkg/parser"
	"gurch101.github.io/go-web/pkg/validation"
)

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
		middleware.UnprocessableEntityResponse(w, r, err)
		return
	}

	v := validation.NewValidator()
	v.Required(createTenantRequest.TenantName, "tenantName", "Tenant Name is required")
	v.Email(createTenantRequest.ContactEmail, "contactEmail", "Contact Email is required")
	v.Check(IsValidTenantPlan(createTenantRequest.Plan), "plan", "Invalid plan")

	if v.HasErrors() {
		middleware.FailedValidationResponse(w, r, v.Errors)
		return
	}

	tenantId, err := CreateTenant(tc.DB, createTenantRequest)
	if err != nil {
		if errors.As(err, &validation.ValidationError{}) {
			middleware.FailedValidationResponse(w, r, []validation.ValidationError{err.(validation.ValidationError)})
			return
		}
		middleware.ServerErrorResponse(w, r, err)
		return
	}

	parser.WriteJSON(w, http.StatusCreated, map[string]any{"id": tenantId}, nil)
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

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	db, err := sql.Open("sqlite3", "./app.db?_foreign_keys=1&_journal=WAL")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	tenantController := NewTenantController(db)
	server := &http.Server{
		Addr:         ":8080",
		Handler:      tenantController.GetMux(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	server.ListenAndServe()
}
