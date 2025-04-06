# Routing

`gowebutils` uses the `chi` router to provide a flexible and powerful routing system for your web applications.

### Route Registration Methods

#### Public Routes

```go
// Accessible without authentication
app.AddPublicRoute("GET", "/api/health", healthController.CheckHealth)
```

Public routes can be accessed by anyone, regardless of authentication status. These are suitable for endpoints like health checks, login pages, and public API resources.

#### Protected Routes

```go
// Requires a valid user session
app.AddProtectedRoute("GET", "/api/profile", profileController.GetProfile)
```

Protected routes ensure that only authenticated users can access the endpoint. If a request lacks a valid session, it will be rejected with a 401 Unauthorized response.

#### Permission-Based Routes

```go
// Requires a valid session AND the "MANAGE_USERS" permission
app.AddProtectedRouteWithPermission(
    "POST",
    "/api/users",
    usersController.CreateUser,
    "MANAGE_USERS"
)
```

Permission-based routes add an additional layer of security by checking that the authenticated user has the specified permission. If the user lacks the required permission, the request will be rejected with a 403 Forbidden response.

### Route Middleware

Each route type automatically applies appropriate middleware:

- Protected routes: Include session validation middleware
- Permission-based routes: Include session validation and permission checking middleware
- Public routes: Include basic request processing middleware

For more details on the specific middleware applied to each route type, see the middleware documentation.

By using these route registration methods, you can ensure consistent security policies across your application with minimal boilerplate code.
