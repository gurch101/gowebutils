package generator_test

import (
	"testing"

	"github.com/gurch101/gowebutils/pkg/collectionutils"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/fsutils"
	"github.com/gurch101/gowebutils/pkg/generator"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestParseSchema(t *testing.T) {
	t.Parallel()
	db := testutils.SetupTestDB(t)

	defer fsutils.CloseAndPanic(db)

	pool := dbutils.FromDB(db)

	// in-memory DBs don't appear to have unique indexes, so we'll skip this test for now
	tables, err := generator.ParseSchema(pool)

	if err != nil {
		t.Fatalf("Error parsing schema: %v", err)
	}

	if len(tables) == 0 {
		t.Error("Expected at least one table, but got none")
	}

	usersTable, ok := collectionutils.FindFirst(tables, func(table generator.Table) bool {
		return table.Name == "users"
	})

	if !ok {
		t.Error("Expected to find users table, but it was not found")
	}

	if len(usersTable.Fields) == 0 {
		t.Error("Expected users table to have fields, but it was empty")
	}

	if usersTable.Fields[0].Name != "id" {
		t.Errorf("Expected first field of users table to be 'id', but got '%s'", usersTable.Fields[0].Name)
	}

	if usersTable.Fields[0].DataType.GoType() != "int64" {
		t.Errorf("Expected first field of users table to be of type 'int64', but got '%s'", usersTable.Fields[0].DataType.GoType())
	}

	if usersTable.Fields[0].Constraints[0] != "PRIMARY KEY" {
		t.Errorf("Expected first field of users table to have PRIMARY KEY constraint, but got '%s'", usersTable.Fields[0].Constraints[0])
	}
}
