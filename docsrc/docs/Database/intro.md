---
sidebar_position: 1
---

# Intro

`gowebutils` is optimized for SQLite databases, providing a simple yet powerful interface for database operations.

### Connection Pooling

When initializing your App, the following connection pools are created:

- 1 connection for write operations
- 10 connections for read operations

All connections are configured with foreign keys and WAL (Write-Ahead Logging) mode enabled for improved performance and data integrity.

### Database Operations

Database Operations
The App.DB object provides an interface similar to Go's standard sql.DB with the following methods:

- ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
- QueryContext(ctx context.Context, query string, args ...interface{}) (\*sql.Rows, error)
- QueryRowContext(ctx context.Context, query string, args ...interface{}) \*sql.Row

For write operations, use the app.DB.WithTransaction function to execute your queries within a transaction.

### Migrations

Database migrations are managed using `go-migrate`. Simply add the following to your `Makefile`:

```sh
# Create a new migration in db/migrations
migrate/new:
	migrate create -seq -ext sql -dir db/migrations ${name}

# apply all migrations
migrate/up:
	@migrate -path db/migrations -database sqlite3://${DB_FILEPATH} up

# rollback the last migration
migrate/down:
	@migrate -path db/migrations -database sqlite3://${DB_FILEPATH} down 1
```

### Configuration

A single environment variable is required to initialize your database.

```sh
# The sqlite3 database file path
export DB_FILEPATH="./app.db"
```
