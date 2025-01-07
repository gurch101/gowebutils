package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"gurch101.github.io/go-web/pkg/authutils"
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

func (c *TenantController) RegisterRoutes(router *httputils.Router) {
	router.AddAuthenticatedRoute("POST /tenants", c.CreateTenantHandler)
	router.AddAuthenticatedRoute("GET /tenants/{id}", c.GetTenantHandler)
	router.AddAuthenticatedRoute("GET /tenants", c.SearchTenantsHandler)
	router.AddAuthenticatedRoute("PATCH /tenants/{id}", c.UpdateTenantHandler)
	router.AddAuthenticatedRoute("DELETE /tenants/{id}", c.DeleteTenantHandler)
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

const (
	tenantNameRequestKey   = "tenantName"
	planRequestKey         = "plan"
	contactEmailRequestKey = "contactEmail"
	tenantResourceKey      = "tenants"
)

type CreateTenantRequest struct {
	TenantName   string     `json:"tenantName"`
	ContactEmail string     `json:"contactEmail"`
	Plan         TenantPlan `json:"plan"`
}

func (tc *TenantController) CreateTenantHandler(w http.ResponseWriter, r *http.Request) {
	createTenantRequest := &CreateTenantRequest{}
	err := httputils.ReadJSON(w, r, createTenantRequest)
	if err != nil {
		httputils.UnprocessableEntityResponse(w, r, err)
		return
	}

	v := validation.NewValidator()
	v.Required(createTenantRequest.TenantName, tenantNameRequestKey, "Tenant Name is required")
	v.Email(createTenantRequest.ContactEmail, contactEmailRequestKey, "Contact Email is required")
	v.Check(IsValidTenantPlan(createTenantRequest.Plan), planRequestKey, "Invalid plan")

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
	err = httputils.WriteJSON(w, http.StatusCreated, envelope{"id": tenantId}, headers)
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

	err = httputils.WriteJSON(w, http.StatusOK, &GetTenantResponse{ID: tenant.ID, TenantName: tenant.TenantName, ContactEmail: tenant.ContactEmail, Plan: tenant.Plan, IsActive: tenant.IsActive}, nil)
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
	err = httputils.ReadJSON(w, r, updateTenantRequest)
	if err != nil {
		httputils.UnprocessableEntityResponse(w, r, err)
		return
	}
	tenant.TenantName = validation.Coalesce(updateTenantRequest.TenantName, tenant.TenantName)
	tenant.ContactEmail = validation.Coalesce(updateTenantRequest.ContactEmail, tenant.ContactEmail)
	tenant.Plan = validation.Coalesce(updateTenantRequest.Plan, tenant.Plan)
	tenant.IsActive = validation.Coalesce(updateTenantRequest.IsActive, tenant.IsActive)

	v := validation.NewValidator()
	v.Required(tenant.TenantName, tenantNameRequestKey, "Tenant Name is required")
	v.Email(tenant.ContactEmail, contactEmailRequestKey, "Contact Email is required")
	v.Check(IsValidTenantPlan(tenant.Plan), planRequestKey, "Invalid plan")

	if v.HasErrors() {
		httputils.FailedValidationResponse(w, r, v.Errors)
		return
	}

	err = UpdateTenant(tc.DB, tenant)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)
		return
	}

	err = httputils.WriteJSON(w, http.StatusOK, &GetTenantResponse{ID: tenant.ID, TenantName: tenant.TenantName, ContactEmail: tenant.ContactEmail, Plan: tenant.Plan, IsActive: tenant.IsActive}, nil)
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

	err = httputils.WriteJSON(w, http.StatusOK, envelope{"message": "Tenant successfully deleted"}, nil)
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
	searchTenantsRequest.TenantName = parser.ParseString(qs, tenantNameRequestKey, nil)
	searchTenantsRequest.Plan = parser.ParseString(qs, planRequestKey, nil)
	searchTenantsRequest.IsActive = parser.ParseBool(qs, "isActive", nil)
	searchTenantsRequest.ContactEmail = parser.ParseString(qs, contactEmailRequestKey, nil)
	searchTenantsRequest.ParseFilters(qs, v, []string{"id", tenantNameRequestKey, planRequestKey, contactEmailRequestKey, fmt.Sprintf("-%s", tenantNameRequestKey), fmt.Sprintf("-%s", planRequestKey), fmt.Sprintf("-%s", contactEmailRequestKey)})
	if v.HasErrors() {
		httputils.FailedValidationResponse(w, r, v.Errors)
		return
	}

	tenants, pagination, err := SearchTenants(tc.DB, searchTenantsRequest)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)
		return
	}
	err = httputils.WriteJSON(w, http.StatusOK, envelope{"metadata": pagination, tenantResourceKey: tenants}, nil)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

var ErrTenantAlreadyRegistered = validation.Error{
	Field:   tenantNameRequestKey,
	Message: "This tenant is already registered",
}

// service layer
func CreateTenant(db *sql.DB, createTenantRequest *CreateTenantRequest) (*int64, error) {
	tenantModel := NewTenantModel(createTenantRequest.TenantName, createTenantRequest.ContactEmail, createTenantRequest.Plan)

	id, err := InsertTenant(db, tenantModel)

	if err != nil {
		if errors.Is(err, dbutils.ErrUniqueConstraint) && strings.Contains(err.Error(), tenantNameDbFieldName) {
			return nil, ErrTenantAlreadyRegistered
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

type User struct {
	ID int64
	TenantID int64
	UserName string
	Email string
}

type AuthService struct {
	db *sql.DB
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

func (a *AuthService) GetOrCreateUser(ctx context.Context, username, email string) (User, error) {
	return User{
		ID: 1,
		TenantID: 1,
		UserName: username,
		Email: email,
	}, nil
}

func (a* AuthService) UserExists(ctx context.Context, user User) (bool, error) {
	return true, nil
}

func main() {
	logger := httputils.InitializeSlog(parser.ParseEnvString("LOG_LEVEL", "info"))

	db, err := sql.Open(dbutils.SqliteDriverName, fmt.Sprintf("%s?_foreign_keys=1&_journal=WAL", parser.ParseEnvStringPanic("DB_FILEPATH")))

	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}

	defer func() {
		closeErr := db.Close()
		if closeErr != nil {
			panic(closeErr)
		}
	}()

	oidcConfig, err := authutils.CreateOauthConfig(
		parser.ParseEnvStringPanic("OIDC_CLIENT_ID"),
		parser.ParseEnvStringPanic("OIDC_CLIENT_SECRET"),
		parser.ParseEnvStringPanic("OIDC_DISCOVERY_URL"),
		parser.ParseEnvStringPanic("OIDC_ISSUER_URL"),
		parser.ParseEnvStringPanic("OIDC_SP_URL"),
	)

	if err != nil {
		panic(err)
	}

	allowedOrigins := strings.Fields(parser.ParseEnvString("CORS_ALLOWED_ORIGINS", ""))
	fileServer := http.FileServer(http.Dir("./web/static/"))
	sessionManager := authutils.CreateSessionManager(db)
	authService := NewAuthService(db)

	oidcController := authutils.NewOidcController(sessionManager, authService.GetOrCreateUser, oidcConfig)
	router := httputils.NewRouter(
		authutils.GetSessionMiddleware(sessionManager, authService.UserExists),
	)
	router.Use(
		httputils.LoggingMiddleware,
		httputils.RecoveryMiddleware,
		httputils.RateLimitMiddleware,
		sessionManager.LoadAndSave,
	)

	tenantController := NewTenantController(db)
	tenantController.RegisterRoutes(router)
	oidcController.RegisterRoutes(router)
	router.AddStaticRoute("/static/", httputils.GetCORSMiddleware(allowedOrigins)(http.StripPrefix("/static", fileServer)))

	err = httputils.ServeHTTP(router.BuildHandler(), logger)

	if err != nil {
		slog.Error(err.Error())
	}
}
