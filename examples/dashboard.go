package main

import (
	"net/http"

	"github.com/gurch101/gowebutils/pkg/app"
	"github.com/gurch101/gowebutils/pkg/httputils"
)

type DashboardController struct {
	app *app.App
}

func NewDashboardController(app *app.App) *DashboardController {
	return &DashboardController{app: app}
}

func (c *DashboardController) Dashboard(w http.ResponseWriter, r *http.Request) {
	err := c.app.RenderTemplate(w, "index.go.tmpl", nil)

	if err != nil {
		httputils.ServerErrorResponse(w, r, err)
	}
}
