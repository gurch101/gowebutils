---
sidebar_position: 2
---

# CRUD Helpers

[godoc](https://pkg.go.dev/github.com/gurch101/gowebutils/pkg/dbutils)

The `dbutils` package provides a set of helper functions for performing CRUD (Create, Read, Update, Delete) operations on single records. The only requirement for these helpers is that the datable tables need to have a unique numeric `id` column and a numeric `version` column.

## Usage

### Create

The `Insert` function inserts a single record into a table. It takes a context, a database connection, a table name, and a map of column names and values. It returns the ID of the inserted record and an error if any.

```go
type User struct {
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

### Read

There are a few ways to read a record from the database.

#### Get By ID

The `GetByID` function retrieves a single record from a table by its ID. It takes a context, a database connection, a table name, and an id. It returns a pointer to the record and an error, if any.

#### Get By

The `GetBy` function retrieves a single record from a table by a where clause. It takes a context, a database connection, a table name, and a where clause. It returns a pointer to the record and an error, if any.

#### Exists

The `Exists` function checks if a record exists in a table by a where clause. It takes a context, a database connection, a table name, and a where clause. It returns a boolean and an error, if any.

### Update

The `Update` function updates a single record in a table. It takes a context, a database connection, a table name, an id and version, and a map of column names and values to update. It returns an error, if any. If the version does not match, a `ErrEditConflict` error is returned.

```go
func UpdateUser(ctx context.Context, db dbutils.DB, user *User) error {
  return dbutils.UpdateByID(ctx, db, "users", user.ID, user.Version, map[string]any{
    "name":   user.Name,
    "email": user.Email,
  })
}
```

### Delete

The `Delete` function deletes a single record from a table. It takes a context, a database connection, a table name, and an id. It returns an error, if any.

```go
func DeleteTenantByID(ctx context.Context, db dbutils.DB, userID int64) error {
  return dbutils.DeleteByID(ctx, db, "users", userID)
}
```

## Error Handling

All sql errors returned by the `dbutils` helper functions are wrapped. See [errors](https://pkg.go.dev/github.com/gurch101/gowebutils/pkg/dbutils#pkg-variables) for all possible errors. Use `errors.Is` to check for specific errors and handle them accordingly.

### Example

```go
if errors.Is(err, dbutils.ErrUniqueConstraint) && strings.Contains(err.Error(), "name") {
  return nil, ErrUserAlreadyRegistered
}
```
