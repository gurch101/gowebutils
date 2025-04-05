---
sidebar_position: 4
---

# Transactions

To use transactions, use `app.DB.WithTransaction` to wrap your database operations in a transaction. Transactions can be nested simply by calling `WithTransaction` again. If an error is returned by your callback function, the transaction will be rolled back.

```go
err := dbutils.WithTransaction(ctx, db, func(tx dbutils.DB) error {
  tenantID, err := dbutils.Insert(ctx, tx, "tenants", map[string]any{
    "tenant_name":   uuid.New().String(),
    "contact_email": email,
    "plan":          "free",
  })

  if err != nil {
    return fmt.Errorf("failed to create tenant: %w", err)
  }

  userID, err = dbutils.Insert(ctx, tx, "users", map[string]any{
    "tenant_id": tenantID,
    "user_name": username,
    "email":     email,
  })

  if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
  }

  return nil
})
```
