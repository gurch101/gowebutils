package dbutils_test

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestQueryBuilder_SimpleSelect(t *testing.T) {
	t.Parallel()

	qb := dbutils.NewQueryBuilder(nil).From("users")
	query, args := qb.Build()
	expectedQuery := "SELECT * FROM users"

	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 0 {
		t.Errorf("Expected no args, got %v", args)
	}
}

func TestQueryBuilder_SelectWithFields(t *testing.T) {
	t.Parallel()

	qb := dbutils.NewQueryBuilder(nil).Select("id", "name").From("users")
	query, args := qb.Build()

	expectedQuery := "SELECT id, name FROM users"
	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 0 {
		t.Errorf("Expected no args, got %v", args)
	}
}

func TestQueryBuilder_SelectWithWhere(t *testing.T) {
	t.Parallel()

	qb := dbutils.NewQueryBuilder(nil).Select("id", "name").From("users").Where("id = ?", 1)
	query, args := qb.Build()

	expectedQuery := "SELECT id, name FROM users WHERE (id = ?)"
	expectedArgs := []interface{}{1}

	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}

	if !reflect.DeepEqual(args, expectedArgs) {
		t.Errorf("Expected args %v, got %v", expectedArgs, args)
	}
}

func TestQueryBuilder_SelectWithJoin(t *testing.T) {
	t.Parallel()

	qb := dbutils.NewQueryBuilder(nil).
		Select("u.id", "u.name", "p.name").
		From("users u").
		Join("INNER", "profiles p", "u.id = p.user_id")
	query, args := qb.Build()

	expectedQuery := "SELECT u.id, u.name, p.name FROM users u INNER JOIN profiles p ON u.id = p.user_id"
	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 0 {
		t.Errorf("Expected no args, got %v", args)
	}
}

func TestQueryBuilder_SelectWithGroupBy(t *testing.T) {
	t.Parallel()

	qb := dbutils.NewQueryBuilder(nil).Select("name", "COUNT(*)").From("users").GroupBy("name")
	query, args := qb.Build()

	expectedQuery := "SELECT name, COUNT(*) FROM users GROUP BY name"
	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 0 {
		t.Errorf("Expected no args, got %v", args)
	}
}

func TestQueryBuilder_SelectWithOrderBy(t *testing.T) {
	t.Parallel()

	qb := dbutils.NewQueryBuilder(nil).Select("id", "name").From("users").OrderBy("name", "-id")
	query, args := qb.Build()

	expectedQuery := "SELECT id, name FROM users ORDER BY name ASC, id DESC"
	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 0 {
		t.Errorf("Expected no args, got %v", args)
	}
}

func TestQueryBuilder_SelectWithLimitAndOffset(t *testing.T) {
	t.Parallel()

	qb := dbutils.NewQueryBuilder(nil).Select("id", "name").From("users").Limit(10).Offset(20)
	query, args := qb.Build()

	expectedQuery := "SELECT id, name FROM users LIMIT 10 OFFSET 20"
	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 0 {
		t.Errorf("Expected no args, got %v", args)
	}
}

func TestQueryBuilder_SelectWithLike(t *testing.T) {
	t.Parallel()

	filterVal := "doe"
	queryBuilder := dbutils.NewQueryBuilder(nil).Select("id", "name").
		From("users").
		WhereLike("name", dbutils.OpContains, &filterVal).
		AndWhereLike("email", dbutils.OpStartsWith, &filterVal).
		OrWhereLike("phone", dbutils.OpEndsWith, &filterVal)

	query, args := queryBuilder.Build()

	expectedQuery := "SELECT id, name FROM users WHERE (name LIKE ?) AND (email LIKE ?) OR (phone LIKE ?)"
	expectedArgs := []interface{}{"%doe%", "doe%", "%doe"}

	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}

	if !reflect.DeepEqual(args, expectedArgs) {
		t.Errorf("Expected args %v, got %v", expectedArgs, args)
	}
}

func TestQueryBuilder_SelectWithNullWhere(t *testing.T) {
	qb := dbutils.NewQueryBuilder(nil).Select("id", "name").From("users").Where("id = ?", nil).AndWhere("name = ?", "foo")
	query, args := qb.Build()
	expectedQuery := "SELECT id, name FROM users WHERE (name = ?)"

	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %v", args)
	}
}

func TestQueryBuilder_SelectWithPage(t *testing.T) {
	t.Parallel()

	qb := dbutils.NewQueryBuilder(nil).Select("id", "name").From("users").Page(2, 10)
	query, _ := qb.Build()
	expectedQuery := "SELECT id, name FROM users LIMIT 10 OFFSET 10"

	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}
}

func TestQueryBuilder_ComplexSelect(t *testing.T) {
	t.Parallel()

	queryBuilder := dbutils.NewQueryBuilder(nil).
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
	query, args := queryBuilder.Build()

	expectedQuery := "SELECT u.id, u.name, p.name FROM users " +
		"u INNER JOIN profiles p ON u.id = p.user_id " +
		"WHERE (u.age > ?) AND (p.active = ?) OR (u.name LIKE ?) " +
		"GROUP BY u.id, u.name, p.name ORDER BY u.name ASC, u.id DESC LIMIT 10 OFFSET 20"

	expectedArgs := []interface{}{18, true, "%doe%"}

	if query != expectedQuery {
		t.Errorf("Expected query %q, got %q", expectedQuery, query)
	}

	if !reflect.DeepEqual(args, expectedArgs) {
		t.Errorf("Expected args %v, got %v", expectedArgs, args)
	}
}

func TestQueryBuilder_Execute(t *testing.T) {
	t.Parallel()

	type User struct {
		ID   int64
		Name string
	}

	var users []User

	db := testutils.SetupTestDB(t)

	defer func() {
		closeErr := db.Close()
		if closeErr != nil {
			t.Fatalf("Failed to close database connection: %v", closeErr)
		}
	}()

	qb := dbutils.NewQueryBuilder(db).Select("id", "user_name").From("users")

	err := qb.Exec(func(rows *sql.Rows) error {
		var user User

		err := rows.Scan(&user.ID, &user.Name)
		if err != nil {
			return err //nolint: wrapcheck
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
}

func TestQueryBuilder_QueryRow(t *testing.T) {
	t.Parallel()
	db := testutils.SetupTestDB(t)

	defer func() {
		closeErr := db.Close()
		if closeErr != nil {
			t.Fatalf("Failed to close database connection: %v", closeErr)
		}
	}()

	var id int64

	err := dbutils.NewQueryBuilder(db).Select("id").From("users").Where("id = ?", 1).QueryRow(&id)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if id != 1 {
		t.Errorf("Expected id 1, got %d", id)
	}
}

func TestQueryBuilder_QueryRowNoRow(t *testing.T) {
	t.Parallel()

	db := testutils.SetupTestDB(t)
	defer func() {
		closeErr := db.Close()
		if closeErr != nil {
			t.Fatalf("Failed to close database connection: %v", closeErr)
		}
	}()

	var id int64

	err := dbutils.NewQueryBuilder(db).Select("id").From("users").Where("id = ?", 999).QueryRow(&id)
	if !errors.Is(err, dbutils.ErrRecordNotFound) {
		t.Errorf("Expected no record error, got %v", err)
	}
}
