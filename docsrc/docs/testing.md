# Testing

`gowebutils` encourages end-to-end testing of your application from handlers to live database queries. To bootstrap your application for testing, you can use helpers in the `testutils` package:

```go
func TestCreateTenant(t *testing.T) {
  t.Parallel()
  app := testutils.NewTestApp(t)
  defer app.Close()

  createTenantController := NewCreateTenantController(app.App)
  app.TestRouter.Post("/tenants", createTenantController.CreateTenantHandler)

  // Define the input JSON for the request
  createTenantRequest := map[string]interface{}{
    "tenantName":   "TestTenant",
    "contactEmail": "test@example.com",
    "plan":         "free",
  }

  req := testutils.CreatePostRequest(t, "/tenants", createTenantRequest)
  rr := app.MakeRequest(req)

  // Make assertions...
}
```
