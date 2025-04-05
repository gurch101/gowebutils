---
sidebar_position: 1
---

# Intro

`gowebutils` is optimized for use with SQLite databases. When initializing your `App`, a connection pool with 1 connection is created that supports writes and 10 connections are created that support reads. When issuing writes, simply use the `App.DB.WithTransaction` function to execute your write queries. Connections are created with foreign keys and WAL mode enabled.

The `App.DB` object shares a similar interface to the `sql.DB` object and has the following methods: - `ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)` - `QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)` - `QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row`

### Intialization

A single environment variable is required to initialize your database.

```sh
# The sqlite3 database file path
export DB_FILEPATH="./app.db"
```
