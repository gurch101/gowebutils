package main

import (
	"embed"
	"log/slog"

	"github.com/gurch101/gowebutils/pkg/app"

	// needed for sqlite3 driver.
	_ "github.com/mattn/go-sqlite3"
)

type envelope map[string]interface{}

/*
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

	location, err := c.app.FileService.UploadFile(handler.Filename, file)
	if err != nil {
		fmt.Println("Error uploading file:", err)
		return
	}

	slog.Info("upload file", "location", location)
}

func (c *TenantController) DownloadFile(w http.ResponseWriter, r *http.Request) {
	contents, err := c.app.FileService.DownloadFile("1.pdf")
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

	err := c.app.FileService.DeleteFile("1.pdf")
	if err != nil {
		fmt.Println("Error deleting file:", err)
		return
	}

	w.WriteHeader(http.StatusOK)

	slog.Info("delete file")
}
*/

//go:embed templates/email
var emailTemplates embed.FS

//go:embed templates/html
var htmlTemplates embed.FS

func main() {
	app, err := app.NewApp(
		app.WithEmailTemplates(emailTemplates),
		app.WithHTMLTemplates(htmlTemplates),
		app.WithGetUserExistsFn(GetUserExists),
		app.WithGetOrCreateUserFn(GetOrCreateUser),
	)

	if err != nil {
		slog.Error(err.Error())

		return
	}

	defer app.Close()

	createTenantController := NewCreateTenantController(app)
	app.AddProtectedRoute("POST", "/api/tenants", createTenantController.CreateTenantHandler)
	getTenantController := NewGetTenantController(app)
	app.AddProtectedRoute("GET", "/api/tenants/:id", getTenantController.GetTenantHandler)
	deleteTenantController := NewDeleteTenantController(app)
	app.AddProtectedRoute("DELETE", "/api/tenants/:id", deleteTenantController.DeleteTenantHandler)
	updateTenantController := NewUpdateTenantController(app)
	app.AddProtectedRoute("PUT", "/api/tenants/:id", updateTenantController.UpdateTenantHandler)
	searchTenantController := NewSearchTenantController(app)
	app.AddProtectedRoute("GET", "/api/tenants", searchTenantController.SearchTenantsHandler)
	dashboardController := NewDashboardController(app)
	app.AddProtectedRoute("GET", "/", dashboardController.Dashboard)
	inviteUserController := NewInviteUserController(app)
	app.AddProtectedRoute("POST", "/api/invite", inviteUserController.InviteUser)

	err = app.Start()
	if err != nil {
		slog.Error(err.Error())
	}
}
