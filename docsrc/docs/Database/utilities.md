---
sidebar_position: 2
---

# CRUD Helpers

[godoc](https://pkg.go.dev/github.com/gurch101/gowebutils/pkg/dbutils)

The `dbutils` package provides a set of helper functions for performing CRUD (Create, Read, Update, Delete) operations on single database records. These helpers simplify common database tasks while maintaining proper error handling and data consistency.

### Requirements

To use these helper functions, your database tables must have:

- A unique numeric id column
- A numeric version column for optimistic concurrency control

### Create Operations

#### Insert

The `Insert` function adds a single record to a table and returns the ID of the newly created record.

```go
type User struct {
  ID      int64
  Name   string
  Email  string
  Version int64
}

func InsertUser(ctx context.Context, db dbutil.DB, user *User) (*int64, error) {
  return dbutils.Insert(ctx, db, "users", map[string]any{
    "name":   user.Name,
    "email": user.Email,
  })
}
```

### Read Operations

#### Get By ID

Retrieves a single record by its ID. Returns `dbutils.ErrRecordNotFound` if no matching record exists.

```go
func GetUserByID(ctx context.Context, db dbutils.DB, userid int64) (authutils.User, error) {
  var user User
  err := dbutils.GetByID(ctx, db, "users", userid, map[string]any{
    "id":        &user.ID,
    "name":      &user.Name,
    "email":     &user.Email,
  })

  if err != nil {
    return authutils.User{}, fmt.Errorf("get user query failed: %w", err)
  }

  return user, nil
}
```

#### Get By

Retrieves a single record using a custom WHERE clause. Returns `dbutils.ErrRecordNotFound` if no matching record exists. If multiple records match, only the first one is returned.

```go
func GetUserByEmail(ctx context.Context, db dbutils.DB, email string) (authutils.User, error) {
  var user User
  err := dbutils.GetBy(ctx, db, "users", userid, map[string]any{
    "id":        &user.ID,
    "name":      &user.Name,
    "email":     &user.Email,
  }, map[string]any {
    "email": email,
  })

  if err != nil {
    return authutils.User{}, fmt.Errorf("get user query failed: %w", err)
  }

  return user, nil
}
```

#### Exists

Checks if a record with the specified ID exists in the table.

```go
func GetUserExists(ctx context.Context, db dbutils.DB, id int64) bool {
  return dbutils.Exists(ctx, db, "users", id)
}
```

### Update Operations

#### Update By ID

Updates a record with optimistic concurrency control using the version column. Returns `ErrEditConflict` if the version doesn't match (indicating the record was modified by another process).

```go
func UpdateUser(ctx context.Context, db dbutils.DB, user *User) error {
  return dbutils.UpdateByID(ctx, db, "users", user.ID, user.Version, map[string]any{
    "name":   user.Name,
    "email": user.Email,
  })
}
```

### Delete Operations

#### Delete By ID

Removes a record from the table by its ID.

```go
func DeleteTenantByID(ctx context.Context, db dbutils.DB, userID int64) error {
  return dbutils.DeleteByID(ctx, db, "users", userID)
}
```

## Error Handling

All SQL errors returned by the `dbutils` helper functions are wrapped with additional context. Use `errors.Is` to check for specific error types and handle them appropriately.

```go
if errors.Is(err, dbutils.ErrUniqueConstraint) && strings.Contains(err.Error(), "name") {
  return nil, ErrUserAlreadyRegistered
}
```

See the [package documentation](https://pkg.go.dev/github.com/gurch101/gowebutils/pkg/dbutils#pkg-variables) for a complete list of error types.
