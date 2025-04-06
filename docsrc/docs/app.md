---
sidebar_position: 2
---

# Application

The App struct provides a comprehensive set of methods for building web applications, including:

- Routing
- Session management
- OIDC integration
- Configuration management
- Database connection pooling
- File uploads/downloads with S3
- HTML templating
- Email delivery

All these features are packaged in a single, easy-to-use interface.

### Initialization

```go
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

	err = app.Start()
	if err != nil {
		slog.Error(err.Error())
	}
}
```
