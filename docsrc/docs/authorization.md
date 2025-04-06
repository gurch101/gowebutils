# Authorization

`gowebutils` provides a robust Role-Based Access Control (RBAC) system for managing user permissions in your application.

### Database Schema

Set up your database with the following tables to enable authorization:

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) UNIQUE,
    is_active BOOLEAN DEFAULT true,
    role_id INTEGER REFERENCES roles(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Roles table
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT
);


-- Role-Permission mapping table
CREATE TABLE role_permissions (
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    code VARCHAR(255) NOT NULL UNIQUE,  -- e.g., "create_post", "delete_user"
    description TEXT,
    PRIMARY KEY (role_id, code)
);
```

### Authorization Model

The authorization system follows a simple hierarchy:

- Each user is assigned to a single role
- Each role can have multiple permissions
- A permission is a code and description representing an action that can be performed

### Implementing Authorization

You can check permissions in two ways:

#### Middleware-based Authorization

Protect routes by requiring specific permissions:

```go
// Returns a 403 Forbidden response if the user lacks the CREATE_POST permission
app.AddProtectedRouteWithPermission("POST", "/api/post", handler, "CREATE_POST")
```

This approach automatically checks permissions before the handler is executed, simplifying your route definitions.

#### Handler-based Authorization

Check permissions directly within your handler functions:

```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
  user := GetUserFromContext(r.Context())
  if !user.HasAnyPermission("CREATE_POST", "VIEW_POST", "EDIT_POST") {
    httputils.ForbiddenResponse(w, r)

    return
  }
}
```

This method provides more flexibility when you need complex permission checks or conditional authorization logic.
