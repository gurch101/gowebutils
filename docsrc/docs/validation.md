# Request Validation

`gowebutils` provides a simple yet powerful validation system to ensure your incoming requests meet your application's requirements.

### Basic Usage

Here's how to validate incoming request data:

```go
type CreateTenantRequest struct {
	TenantName   string     `json:"tenantName"`
	ContactEmail string     `json:"contactEmail"`
	Plan         TenantPlan `json:"plan"`
}

func (c *CreateTenantController) CreateTenantHandler(w http.ResponseWriter, r *http.Request) {
  createTenantRequest, err := httputils.ReadJSON[CreateTenantRequest](w, r)
  if err != nil {
    httputils.UnprocessableEntityResponse(w, r, err)

    return
  }

  // Create a new validator instance
  v := validation.NewValidator()

  // Run validation rules
  v.Required(tenant.TenantName, "tenantName", "Tenant Name is required")
  v.Email(tenant.ContactEmail, "contactEmail", "Contact Email is required")
  v.Check(IsValidTenantPlan(tenant.Plan), "plan", "Invalid plan")

  // Check if any validation rules failed
  if v.HasErrors() {
    // Return a 400 Bad Request with validation errors
    httputils.FailedValidationResponse(w, r, v.Errors)

    return
  }

  // Continue processing the valid request...
```

### Validation Error Response Format

When validation fails, `httputils.FailedValidationResponse` returns a 400 status code with a structured JSON response:

```json
{
  "errors": [
    {
      "field": "tenantName",
      "message": "Tenant Name is required"
    },
    {
      "field": "contactEmail",
      "message": "Contact Email is required"
    },
    {
      "field": "plan",
      "message": "Invalid plan"
    }
  ]
}
```
