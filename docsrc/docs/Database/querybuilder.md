---
sidebar_position: 3
---

# Query Builder

While the dbutils CRUD helper functions are useful for simple database operations, they may not be sufficient for complex or dynamic queries. For these scenarios, gowebutils provides a powerful querybuilder that supports:

- Joins
- Complex WHERE clauses
- GROUP BY statements
- ORDER BY clauses
- LIMIT and OFFSET pagination

### Basic Usage

```go
// Executes:
// SELECT id, name, age FROM users WHERE (age > 18) OR (name LIKE "%doe%")
// LIMIT 10 OFFSET 10
querybuilder := dbutils.NewQueryBuilder(db).
  Select("id", "name", "age").
  From("users").
  Where("age > ?", 18).
  OrWhereLike("name", dbutils.OpContains, "doe").
  Limit(10).
  Offset(10).
  Query(func(rows *sql.Rows) error {
    // do something with rows
  })
```

### Advanced Usage

```go
// runs SELECT u.id, COUNT(c.comments) FROM users u
// INNER JOIN comments c ON u.id = c.user_id
// WHERE (u.age > 18) AND (u.active) OR (u.name = "doe")
// GROUP BY u.id ORDER BY u.name DESC LIMIT 10 OFFSET 20
queryBuilder := dbutils.NewQueryBuilder(db).
  Select("u.id", "COUNT(c.comments)").
  From("users u").
  Join("INNER", "comments c", "u.id = c.user_id").
  Where("u.age > ?", 18).
  AndWhere("u.active = ?", true).
  OrWhere("u.name = ?", "doe").
  GroupBy("u.id").
  // -<fieldname> is used for descending order
  // <fieldname> is used for ascending order
  OrderBy("-u.name").
  Limit(10).
  Offset(20).
  Query(func(rows *sql.Rows) error {
    // do something with rows
  })
```

### Handling NULL Values

NULL values passed to any of the WHERE clause functions are automatically ignored. This feature allows you to avoid conditional branching in your code when dealing with optional filter parameters.

### Getting the Raw Query

If you need the raw SQL query and arguments instead of executing it directly, use the Build() method:

```go
	qb := dbutils.NewQueryBuilder(db).Select("id", "name").From("users").Where("id = ?", 1)
	// query = SELECT id, name FROM users WHERE (id = ?)
	// args = []{1}
	query, args := qb.Build()
```

This approach is useful when you need to use the query with other database functions or for debugging purposes.
