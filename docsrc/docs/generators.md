# Generators

`gowebutils` includes a CLI-based generator that scaffolds boilerplate code based on your SQLite database schema. It can quickly create handlers, services, repositories, and tests using your database schema.

### Installation

Install the generator CLI with:

```sh
go install github.com/gurch101/gowebutils/cmd/generator@latest
```

### Getting Started

**1. Set your database file path:**

Make sure the DB_FILEPATH environment variable points to your SQLite .db file.

**2. Run the generator:**

In your project directory, run `generator` in the terminal and select the database tables to scaffold.

### Generated Files

- `internal/<dbtable>/create_<dbtable>.go` - This file contains a handler, service, and repository to create a new record in the database.
- `internal/<dbtable>/get_<dbtable>_by_id.go` - This file contains a handler, service, and repository to get a record from the database by id.
- `internal/<dbtable>/search_<dbtable>.go` - This file contains a handler, service, and repository to search for records in the database by enabling pagination, filter on any field, and sort by any field. You can also control the fields that are returned in the response.
- `internal/<dbtable>/update_<dbtable>.go` - This file contains a handler, service, and repository to update a record in the database via a PATCH request. This file is only generated if your database table has a `version` field.
- `internal/<dbtable>/delete_<dbtable>_by_id.go` - This file contains a handler, service, and repository to delete a record from the database.
- `internal/<dbtable>/<dbtable>_exists.go` - A helper function to check if a record exists in the database.
- `internal/<dbtable>/models.go` - This file contains a struct representing the database table.
- `internal/<dbtable>/test_helpers.go` - This file contains helper functions to create test records for the database table.
- `internal/<dbtable>/routes.go` - This file contains binds the handler functions to the routes through the `app` object.
- `internal/schema_test.go` - This file contains tests to ensure the database schema meets the requirements of the generator.

Each program will have a corresponding end-to-end test file named `internal/<dbtable>/<progname>_test.go`.

### OpenAPI Documentation

Each generated handler includes comments compatible with `swaggo` to automatically generate OpenAPI documentation.

### Code Customization

All generated code is intended as a starting point. You are encouraged to modify it to meet your application's requirements. Generated files should be committed to version control and maintained alongside your application code.

### Schema Conventions

To ensure the generator works optimally, follow these conventions in your database schema:

- Primary Key: Must be named `id` and of type `int64`.

- Table Names: Use _plural_, _snake_case_, and _lowercase_ (e.g., `users`, `order_items`).

- Column Names: Use _snake_case_ and _lowercase_.

- Required Fields: Add `CHECK` constraints like `CHECK (column_name <> '')` to mark fields as required in the generated create/update handlers.

- Email Fields: Any column name containing the word email is assumed to be an email address and will be automatically validated.
