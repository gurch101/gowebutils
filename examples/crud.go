package main

import (
	"database/sql"
	"embed"
	"encoding/gob"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gurch101/gowebutils/pkg/authutils"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/mailutils"
	"github.com/gurch101/gowebutils/pkg/parser"
	"github.com/gurch101/gowebutils/pkg/starter"
	"github.com/gurch101/gowebutils/pkg/templateutils"
	"github.com/gurch101/gowebutils/pkg/validation"

	// needed for sqlite3 driver.
	_ "github.com/mattn/go-sqlite3"
)

type envelope map[string]interface{}

type Controller interface {
	GetMux() *http.ServeMux
}

type TenantController struct {
	DB              *sql.DB
	htmlTemplateMap map[string]*template.Template
}

func NewTenantController(db *sql.DB, htmlTemplateMap map[string]*template.Template) *TenantController {
	return &TenantController{DB: db, htmlTemplateMap: htmlTemplateMap}
}

func (c *TenantController) PublicRoutes(_ httputils.Router) {

}

func (c *TenantController) ProtectedRoutes(router httputils.Router) {
	router.Post("/tenants", c.CreateTenantHandler)
	router.Get("/tenants/{id}", c.GetTenantHandler)
	router.Get("/tenants", c.SearchTenantsHandler)
	router.Patch("/tenants/{id}", c.UpdateTenantHandler)
	router.Delete("/tenants/{id}", c.DeleteTenantHandler)
	router.Post("/api/invite", c.InviteUser)
	router.Get("/", c.Dashboard)
}

type InviteUserRequest struct {
	UserName string `json:"userName"`
	Email    string `json:"email"`
}

func (c *TenantController) InviteUser(w http.ResponseWriter, r *http.Request) {
	inviteUserRequest, err := httputils.ReadJSON[InviteUserRequest](w, r)
	if err != nil {
		httputils.BadRequestResponse(w, r, err)

		return
	}

	user := authutils.ContextGetUser[User](r)

	payload := map[string]any{
		"tenant_id": user.TenantID,
		"email":     inviteUserRequest.Email,
	}

	inviteToken, err := authutils.CreateInviteToken(payload)

	if err != nil {
		httputils.ServerErrorResponse(w, r, err)

		return
	}

	httputils.WriteJSON(w, http.StatusOK, map[string]string{"token": inviteToken}, nil)
}

func (c *TenantController) Dashboard(w http.ResponseWriter, r *http.Request) {
	err := c.htmlTemplateMap["index.go.tmpl"].ExecuteTemplate(w, "index.go.tmpl", nil)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
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
	createTenantRequest, err := httputils.ReadJSON[CreateTenantRequest](w, r)
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

	tenantID, err := CreateTenant(tc.DB, &createTenantRequest)
	if err != nil {
		httputils.HandleErrorResponse(w, r, err)

		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/tenants/%d", tenantID))

	err = httputils.WriteJSON(w, http.StatusCreated, envelope{"id": tenantID}, headers)
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

	updateTenantRequest, err := httputils.ReadJSON[UpdateTenantRequest](w, r)
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

	err = UpdateTenant(r.Context(), tc.DB, tenant)
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
	queryString := r.URL.Query()
	searchTenantsRequest := &SearchTenantsRequest{
		TenantName:   parser.ParseQSString(queryString, tenantNameRequestKey, nil),
		Plan:         parser.ParseQSString(queryString, planRequestKey, nil),
		IsActive:     parser.ParseQSBool(queryString, "isActive", nil),
		ContactEmail: parser.ParseQSString(queryString, contactEmailRequestKey, nil),
	}

	searchTenantsRequest.ParseQSFilters(queryString, v, []string{"id", tenantNameRequestKey, planRequestKey, contactEmailRequestKey, fmt.Sprintf("-%s", tenantNameRequestKey), fmt.Sprintf("-%s", planRequestKey), fmt.Sprintf("-%s", contactEmailRequestKey)})
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

//go:embed templates/email
var emailTemplates embed.FS

//go:embed templates/html
var htmlTemplates embed.FS

func main() {
	db := dbutils.Open(parser.ParseEnvStringPanic("DB_FILEPATH"))

	defer func() {
		closeErr := db.Close()
		if closeErr != nil {
			panic(closeErr)
		}
	}()

	emailTemplateMap := templateutils.LoadTemplates(emailTemplates, "templates/email")

	htmlTemplateMap := templateutils.LoadTemplates(htmlTemplates, "templates/html")

	mailer := mailutils.InitMailer(emailTemplateMap)

	gob.Register(User{})

	authService := NewAuthService(db, mailer, parser.ParseEnvStringPanic("HOST"))
	tenantController := NewTenantController(db, htmlTemplateMap)
	err := starter.CreateAppServer[User](authService, db, tenantController)

	if err != nil {
		slog.Error(err.Error())
	}
}
