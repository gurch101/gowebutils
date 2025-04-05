---
sidebar_position: 3
---

# Query Builder

While `dbutils` CRUD helper functions are useful for simple database operations, they may not be sufficient for complex queries or when you need to build dynamic queries. For these cases, you can use the `querybuilder` which supports joins, complex where clauses, group by, order by, limit, and offset.

## Usage

```go
// runs SELECT id, name, age FROM users WHERE (age > 18) OR (name LIKE "%doe%") LIMIT 10 OFFSET 10
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

```go
// runs SELECT u.id, COUNT(c.comments) FROM users u INNER JOIN comments c ON u.id = c.user_id WHERE (u.age > 18) AND (u.active) OR (u.name = "doe") GROUP BY u.id ORDER BY u.name DESC LIMIT 10 OFFSET 20
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

nils values passed to any of the where clause functions will be ignored which means that you can avoid branching in your code.

Rather than using `Exec`, you can also use the QueryBuilder to get the query string and arguments by calling the `Build()` method.

```go
	qb := dbutils.NewQueryBuilder(db).Select("id", "name").From("users").Where("id = ?", 1)
	// returns SELECT id, name FROM users WHERE (id = ?) and []{1}
	query, args := qb.Build()
```
