---
sidebar_position: 4
---

# Transactions

`gowebutils` provides a simple yet powerful way to manage database transactions, ensuring data consistency across multiple operations.

### Using Transactions

To execute operations within a transaction, use the app.DB.WithTransaction method. This method wraps your database operations in a transaction and handles commit or rollback automatically based on the result of your callback function.

```go
err := app.DB.WithTransaction(ctx, func(tx dbutils.DB) error {
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

  // Transaction will be committed if we return nil
  return nil
})
```

### Nested Transactions

Transactions can be nested by simply calling `WithTransaction` again within a transaction callback. This is particularly useful when composing functions that each need their own transaction context.

### Error Handling

If your callback function returns an error:

1. The transaction will be automatically rolled back
2. No changes will be committed to the database
3. The error will be returned from the WithTransaction call

This approach simplifies error handling and ensures your database remains in a consistent state even when operations fail.
