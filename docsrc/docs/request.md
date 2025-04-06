# Working with Requests and Responses

`gowebutils` builds upon Go's standard `net/http` package, providing helper functions that simplify common request and response handling tasks.

### Request Parsing

#### Path Parameters

Extract values from URL path segments:

```go
import (
  "net/http"
  "github.com/gurch101/gowebutils/pkg/parser"
)

// 1. Define your route with path parameters
app.AddProtectedRoute("GET", "/api/users/:id", usersController.GetUser)

// 2. Parse the path parameters in your handler
func (c *UsersController) GetUser(w http.ResponseWriter, r *http.Request) {
  // Extract the :id parameter as an int64
  // Returns ErrInvalidPathParam if parameter is missing or invalid
  id, err := parser.ParseIDPathParam(r)
  if err != nil {
    // Handle error
  }

  // Continue with handler logic
}
```

#### Query Parameters

Parse URL query string parameters with type conversion and default values:

```go
func (c *UsersController) SearchUsers(w http.ResponseWriter, r *http.Request) {
  queryString := r.URL.Query()

  searchUsersRequest := &SearchTenantsRequest{
    // Get "name" as string (returns nil if not present)
    Name: parser.ParseQSString(queryString, "name", nil),
    // Get "isActive" as bool (returns true if not present)
    IsActive: parser.ParseQSBool(queryString, "isActive", true),
    // Get "numberOfResults" as int (returns 100 if not present)
    NumberOfResults: parser.ParseQSInt(queryString, "numberOfResults", 100),
	}

  // Use the parsed parameters
}
```

#### Request Body

Parse JSON request bodies into typed structs:

```go
type CreateUserRequest struct {
  Name   string    `json:"userName"`
  Email string     `json:"email"`
}

func (c *UsersController) CreateUser(w http.ResponseWriter, r *http.Request) {
  // Parse JSON body into CreateUserRequest struct
  // Returns ErrInvalidJSON if body is invalid
  // Returns error if body contains unknown fields
  createUserRequest, err := httputils.ReadJSON[CreateUserRequest](w, r)
  if err != nil {
    httputils.UnprocessableEntityResponse(w, r, err)

    return
  }

  // Use the parsed request
}
```

### Response Handling

#### JSON Responses

Send structured JSON responses:

```go
func (c *TenantsController) GetUser(w http.ResponseWriter, r *http.Request) {
  // ...
  // Write JSON response with status code and optional headers
  err = httputils.WriteJSON(w, http.StatusOK, &GetTenantResponse{
    ID:     user.ID,
    Name:   user.Name,
    email:  user.Email,
  }, nil)

  if err != nil {
    httputils.ServerErrorResponse(w, r, err)
  }
}
```

#### Error Handling

Send consistent error responses using the provided helper functions:

```go
// Available error response helpers:
httputils.BadRequestResponse(w, r, err)
httputils.EditConflictResponse(w, r)
httputils.FailedValidationResponse(w, r, errors)
httputils.NotFoundResponse(w, r)
httputils.RateLimitExceededResponse(w, r)
httputils.ServerErrorResponse(w, r, err)
httputils.UnauthorizedResponse(w, r)
httputils.UnprocessableEntityResponse(w, r, err)
```

For service-layer errors, use the generic error handler:

```go
// Automatically selects the appropriate error response
// based on the error type
httputils.HandleErrorResponse(w, r, err)
```

### Working with Request Context

#### Request ID

Every request processed by routes registered with `app.AddPublicRoute`, `app.AddProtectedRoute`, or `app.AddProtectedRouteWithPermission` includes a unique `RequestIDKey` in the request context. This ID is automatically included in all `slog` log entries.

### User Information

For protected routes, you can access the authenticated user:

```go
// Get the current user from the request context
user := authutils.GetUserFromContext(request.Context())
```

This provides a convenient way to access user information in your handler functions.
