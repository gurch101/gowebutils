package main

import (
	"embed"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gurch101/gowebutils/pkg/authutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/starter"
	"github.com/gurch101/gowebutils/pkg/templateutils"

	// needed for sqlite3 driver.
	_ "github.com/mattn/go-sqlite3"
)

type envelope map[string]interface{}

type Controller interface {
	GetMux() *http.ServeMux
}

type TenantController struct {
	appserver *starter.AppServer[User]
}

func NewTenantController(appserver *starter.AppServer[User]) *TenantController {
	tenantController := &TenantController{appserver: appserver}
	appserver.AddProtectedRoute("POST", "/tenants", tenantController.createTenantHandler)
	appserver.AddProtectedRoute("GET", "/tenants/{id}", tenantController.getTenantByIdHandler)
	appserver.AddProtectedRoute("GET", "/tenants", tenantController.searchTenantsHandler)
	appserver.AddProtectedRoute("PATCH", "/tenants/{id}", tenantController.updateTenantHandler)
	appserver.AddProtectedRoute("DELETE", "/tenants/{id}", tenantController.deleteTenantHandler)
	appserver.AddProtectedRoute("POST", "/api/invite", tenantController.InviteUser)
	appserver.AddProtectedRoute("GET", "/", tenantController.Dashboard)
	appserver.AddProtectedRoute("POST", "/api/upload", tenantController.UploadFile)
	appserver.AddProtectedRoute("GET", "/api/download/{filename}", tenantController.DownloadFile)
	appserver.AddProtectedRoute("DELETE", "/api/delete/{filename}", tenantController.DeleteFile)
	return tenantController
}

type InviteUserRequest struct {
	UserName string `json:"userName"`
	Email    string `json:"email"`
}

func (c *TenantController) UploadFile(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form, 10 << 20 specifies a maximum upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)

	// Retrieve the file from the form data
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()

	location, err := c.appserver.FileService.UploadFile(handler.Filename, file)
	if err != nil {
		fmt.Println("Error uploading file:", err)
		return
	}

	slog.Info("upload file", "location", location)
}

func (c *TenantController) DownloadFile(w http.ResponseWriter, r *http.Request) {
	contents, err := c.appserver.FileService.DownloadFile("1.pdf")
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=1.pdf")
	w.Header().Set("Content-Type", "application/pdf")
	_, err = w.Write(contents)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}
}

func (c *TenantController) DeleteFile(w http.ResponseWriter, r *http.Request) {

	err := c.appserver.FileService.DeleteFile("1.pdf")
	if err != nil {
		fmt.Println("Error deleting file:", err)
		return
	}

	w.WriteHeader(http.StatusOK)

	slog.Info("delete file")
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
	err := c.appserver.RenderTemplate(w, "index.go.tmpl", nil)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}

//go:embed templates/email
var emailTemplates embed.FS

//go:embed templates/html
var htmlTemplates embed.FS

func main() {
	emailTemplateMap := templateutils.LoadTemplates(emailTemplates, "templates/email")

	htmlTemplateMap := templateutils.LoadTemplates(htmlTemplates, "templates/html")

	appserver := starter.NewAppServer[User](
		htmlTemplateMap,
		emailTemplateMap,
		NewAuthService,
	)

	NewTenantController(appserver)

	err := appserver.Start()

	if err != nil {
		slog.Error(err.Error())
	}
}
