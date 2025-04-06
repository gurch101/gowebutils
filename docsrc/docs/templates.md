# HTML Templates

`gowebutils` provides seamless integration with Go's `html/template` package, allowing you to easily render HTML templates in your web application.

### Setup

Templates are stored in an embedded filesystem and passed to the App during initialization:

```go
//go:embed templates/html
var htmlTemplates embed.FS

func main() {
  app, err := app.NewApp(
    app.WithHTMLTemplates(htmlTemplates),
    // Other configuration options...
  )

  // Continue with app setup
}
```

### Rendering Templates

Once configured, you can render templates from any handler using the `app.RenderTemplate` method:

```go
func (c *DashboardController) Dashboard(w http.ResponseWriter, r *http.Request) {
  // Render the template with the provided data
  // The template path is relative to the embedded filesystem
  err := c.app.RenderTemplate(w, "index.go.tmpl", map[string]string {
    "title": "Dashboard",
  })

  if err != nil {
    httputils.ServerErrorResponse(w, r, err)
  }
}
```
