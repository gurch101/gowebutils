---
sidebar_position: 5
---

# Database Testing

gowebutils simplifies database integration testing by providing utilities to work with real databases in your test environment.

### Setting Up a Test Database

To create an in-memory SQLite database for your tests, use the testutils.SetupTestDB function:

```go
func TestDatabaseOperations(t *testing.T) {
  // Create a test database with migrations and seed data applied
  db := testutils.SetupTestDB(t)
  defer db.Close()

  // Your test code here
}
```

The SetupTestDB function:

1. Creates an in-memory SQLite database
2. Automatically applies all migrations from your db/migrations directory
3. Loads test data from your db/data directory

This approach ensures your tests run against a database with the same schema and constraints as your production environment, while maintaining test isolation and performance.

### Benefits

- Realistic testing: Test against actual database behavior rather than mocks
- Isolation: Each test gets a fresh database instance
- Performance: In-memory SQLite is fast enough for most test suites
- Simplicity: No need to manage external test databases

By using real databases for testing, you can catch issues that might not be apparent when using mocks or stubs.
