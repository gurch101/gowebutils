package generator

//nolint:funlen
func GetSchemaTest() string {
	return `package internal_test

import (
	"strings"
	"testing"

	"github.com/gurch101/gowebutils/pkg/collectionutils"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/generator"
	"github.com/gurch101/gowebutils/pkg/stringutils"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestSchema(t *testing.T) {
	t.Parallel()

	db := testutils.SetupTestDB(t)
	defer db.Close()

	pool := dbutils.FromDB(db)

	schema, err := generator.ParseSchema(pool)

	if err != nil {
		t.Fatal(err)
	}

	if len(schema) == 0 {
		t.Errorf("expected at least one table, got %d", len(schema))
	}

	for _, table := range schema {
		if stringutils.ToSnakeCase(table.Name) != table.Name {
			t.Errorf("expected table name to be snake case, got %s", table.Name)
		}

		for _, field := range table.Fields {
			if stringutils.ToSnakeCase(field.Name) != field.Name {
				t.Errorf("expected column name to be snake case, got %s", field.Name)
			}
		}

		idField, ok := collectionutils.FindFirst(table.Fields, func(field generator.Field) bool {
			return field.Name == "id"
		})

		if !ok {
			t.Errorf("expected table to have an ID field, got %d fields", len(table.Fields))
		}

		if idField.DataType != "Int64" {
			t.Errorf("expected ID field to be an int64, got %s", idField.DataType)
		}

		if !collectionutils.Contains(idField.Constraints, func(constraint string) bool {
			return constraint == "PRIMARY KEY"
		}) {
			t.Errorf("expected ID field to be a primary key, got %s", idField.Constraints)
		}

		createdAtField, ok := collectionutils.FindFirst(table.Fields, func(field generator.Field) bool {
			return field.Name == "created_at"
		})
		if !ok {
			t.Errorf("expected table %s to have a created_at field", table.Name)
		}

		if strings.Join(createdAtField.Constraints, " ") != "NOT NULL DEFAULT CURRENT_TIMESTAMP" {
			t.Errorf("expected created_at field to be NOT NULL DEFAULT CURRENT_TIMESTAMP, got %s", strings.Join(createdAtField.Constraints, " "))
		}

		updatedAtField, hasUpdatedAt := collectionutils.FindFirst(table.Fields, func(field generator.Field) bool {
			return field.Name == "updated_at"
		})

		versionField, hasVersion := collectionutils.FindFirst(table.Fields, func(field generator.Field) bool {
			return field.Name == "version"
		})

		if (hasUpdatedAt && !hasVersion) || (!hasUpdatedAt && hasVersion) {
			t.Errorf("expected table %s to have both updated_at and version, or neither", table.Name)
		}

		if hasUpdatedAt {
			if strings.Join(updatedAtField.Constraints, " ") != "NOT NULL DEFAULT CURRENT_TIMESTAMP" {
				t.Errorf("expected updated_at field to be NOT NULL DEFAULT CURRENT_TIMESTAMP, got %s", strings.Join(updatedAtField.Constraints, " "))
			}
		}

		if hasVersion {
			if versionField.DataType != "Int64" {
				t.Errorf("expected version field to be an int64, got %s", versionField.DataType)
			}

			if !strings.Contains(strings.Join(versionField.Constraints, " "), "NOT NULL DEFAULT 1") {
				t.Errorf("expected version field to be NOT NULL DEFAULT 0, got %s", strings.Join(versionField.Constraints, " "))
			}
		}
	}
}
`
}
