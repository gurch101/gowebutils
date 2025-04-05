---
sidebar_position: 5
---

# Testing

`gowebutils` encourages integration testing through the use of a real database for testing. Simply call

```go
db := testutils.SetupTestDB(t)
defer db.Close()
```

to setup an in-memory SQLite database that has migrations from `db/migrations` applied and data from `db/data` loaded.
