package dbutils

import (
	"database/sql"
	"reflect"
	"testing"
)

func TestQueryBuilder(t *testing.T) {
	t.Run("Build basic SELECT query", func(t *testing.T) {
		qb := NewQueryBuilder(nil).From("users")
		query, args := qb.Build()
		expectedQuery := "SELECT * FROM users"
		if query != expectedQuery {
			t.Errorf("Expected query %q, got %q", expectedQuery, query)
		}
		if len(args) != 0 {
			t.Errorf("Expected no args, got %v", args)
		}
	})

	t.Run("Build SELECT query with fields", func(t *testing.T) {
		qb := NewQueryBuilder(nil).Select("id", "name").From("users")
		query, args := qb.Build()

		expectedQuery := "SELECT id, name FROM users"
		if query != expectedQuery {
			t.Errorf("Expected query %q, got %q", expectedQuery, query)
		}
		if len(args) != 0 {
			t.Errorf("Expected no args, got %v", args)
		}
	})

	t.Run("Build SELECT query with WHERE clause", func(t *testing.T) {
		qb := NewQueryBuilder(nil).Select("id", "name").From("users").Where("id = ?", 1)
		query, args := qb.Build()

		expectedQuery := "SELECT id, name FROM users WHERE (id = ?)"
		expectedArgs := []interface{}{1}
		if query != expectedQuery {
			t.Errorf("Expected query %q, got %q", expectedQuery, query)
		}
		if !reflect.DeepEqual(args, expectedArgs) {
			t.Errorf("Expected args %v, got %v", expectedArgs, args)
		}
	})

	t.Run("Build SELECT query with JOIN clause", func(t *testing.T) {
		qb := NewQueryBuilder(nil).Select("u.id", "u.name", "p.name").From("users u").Join("INNER", "profiles p", "u.id = p.user_id")
		query, args := qb.Build()

		expectedQuery := "SELECT u.id, u.name, p.name FROM users u INNER JOIN profiles p ON u.id = p.user_id"
		if query != expectedQuery {
			t.Errorf("Expected query %q, got %q", expectedQuery, query)
		}
		if len(args) != 0 {
			t.Errorf("Expected no args, got %v", args)
		}
	})

	t.Run("Build SELECT query with GROUP BY clause", func(t *testing.T) {
		qb := NewQueryBuilder(nil).Select("name", "COUNT(*)").From("users").GroupBy("name")
		query, args := qb.Build()

		expectedQuery := "SELECT name, COUNT(*) FROM users GROUP BY name"
		if query != expectedQuery {
			t.Errorf("Expected query %q, got %q", expectedQuery, query)
		}
		if len(args) != 0 {
			t.Errorf("Expected no args, got %v", args)
		}
	})

	t.Run("Build SELECT query with ORDER BY clause", func(t *testing.T) {
		qb := NewQueryBuilder(nil).Select("id", "name").From("users").OrderBy("name", "-id")
		query, args := qb.Build()

		expectedQuery := "SELECT id, name FROM users ORDER BY name ASC, id DESC"
		if query != expectedQuery {
			t.Errorf("Expected query %q, got %q", expectedQuery, query)
		}
		if len(args) != 0 {
			t.Errorf("Expected no args, got %v", args)
		}
	})

	t.Run("Build SELECT query with LIMIT and OFFSET", func(t *testing.T) {
		qb := NewQueryBuilder(nil).Select("id", "name").From("users").Limit(10).Offset(20)
		query, args := qb.Build()

		expectedQuery := "SELECT id, name FROM users LIMIT 10 OFFSET 20"
		if query != expectedQuery {
			t.Errorf("Expected query %q, got %q", expectedQuery, query)
		}
		if len(args) != 0 {
			t.Errorf("Expected no args, got %v", args)
		}
	})

	t.Run("Build SELECT query with LIKE clause", func(t *testing.T) {
		var filterVal = "doe"
		qb := NewQueryBuilder(nil).Select("id", "name").From("users").WhereLike("name", OpContains, &filterVal).AndWhereLike("email", OpStartsWith, &filterVal).OrWhereLike("phone", OpEndsWith, &filterVal)
		query, args := qb.Build()

		expectedQuery := "SELECT id, name FROM users WHERE (name LIKE ?) AND (email LIKE ?) OR (phone LIKE ?)"
		expectedArgs := []interface{}{"%doe%", "doe%", "%doe"}
		if query != expectedQuery {
			t.Errorf("Expected query %q, got %q", expectedQuery, query)
		}
		if !reflect.DeepEqual(args, expectedArgs) {
			t.Errorf("Expected args %v, got %v", expectedArgs, args)
		}
	})

	t.Run("Build SELECT query with nil WHERE clause value", func(t *testing.T) {
		qb := NewQueryBuilder(nil).Select("id", "name").From("users").Where("id = ?", nil).AndWhere("name = ?", "foo")
		query, args := qb.Build()
		expectedQuery := "SELECT id, name FROM users WHERE (name = ?)"
		if query != expectedQuery {
			t.Errorf("Expected query %q, got %q", expectedQuery, query)
		}
		if len(args) != 1 {
			t.Errorf("Expected 1 arg, got %v", args)
		}
	})

	t.Run("Build SELECT query with Page", func(t *testing.T) {
		qb := NewQueryBuilder(nil).Select("id", "name").From("users").Page(2, 10)
		query, _ := qb.Build()
		expectedQuery := "SELECT id, name FROM users LIMIT 10 OFFSET 10"
		if query != expectedQuery {
			t.Errorf("Expected query %q, got %q", expectedQuery, query)
		}
	})

	t.Run("Build complex SELECT query", func(t *testing.T) {
		qb := NewQueryBuilder(nil).
			Select("u.id", "u.name", "p.name").
			From("users u").
			Join("INNER", "profiles p", "u.id = p.user_id").
			Where("u.age > ?", 18).
			AndWhere("p.active = ?", true).
			OrWhere("u.name LIKE ?", "%doe%").
			GroupBy("u.id", "u.name", "p.name").
			OrderBy("u.name", "-u.id").
			Limit(10).
			Offset(20)
		query, args := qb.Build()

		expectedQuery := "SELECT u.id, u.name, p.name FROM users u INNER JOIN profiles p ON u.id = p.user_id WHERE (u.age > ?) AND (p.active = ?) OR (u.name LIKE ?) GROUP BY u.id, u.name, p.name ORDER BY u.name ASC, u.id DESC LIMIT 10 OFFSET 20"
		expectedArgs := []interface{}{18, true, "%doe%"}
		if query != expectedQuery {
			t.Errorf("Expected query %q, got %q", expectedQuery, query)
		}
		if !reflect.DeepEqual(args, expectedArgs) {
			t.Errorf("Expected args %v, got %v", expectedArgs, args)
		}
	})

	t.Run("Execute SELECT query", func(t *testing.T) {
		type User struct {
			ID   int64
			Name string
		}
		var users []User
		db := SetupTestDB(t)
		qb := NewQueryBuilder(db).Select("id", "user_name").From("users")
		err := qb.Execute(func(rows *sql.Rows) error {
			var user User
			err := rows.Scan(&user.ID, &user.Name)
			if err != nil {
				return err
			}
			users = append(users, user)
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(users) != 2 {
			t.Errorf("Expected 2 users, got %d", len(users))
		}
	})
}
