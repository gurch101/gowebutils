package generator_test

import (
	"testing"

	"github.com/gurch101/gowebutils/pkg/generator"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func getTestUserSchema() generator.Table {
	return generator.Table{
		Name: "users",
		Fields: []generator.Field{
			{
				Name:        "id",
				DataType:    generator.SQLInt64,
				Constraints: []string{},
			},
			{
				Name:        "version",
				DataType:    generator.SQLInt64,
				Constraints: []string{},
			},
			{
				Name:        "name",
				DataType:    generator.SQLString,
				Constraints: []string{"UNIQUE", "CHECK (name <> '')"},
			},
			{
				Name:        "email",
				DataType:    generator.SQLString,
				Constraints: []string{"UNIQUE"},
			},
			{
				Name:        "some_int64",
				DataType:    generator.SQLInt64,
				Constraints: []string{},
			},
			{
				Name:        "tenant_id",
				DataType:    generator.SQLInt64,
				Constraints: []string{"NOT NULL"},
			},
			{
				Name:        "some_bool",
				DataType:    generator.SQLBoolean,
				Constraints: []string{},
			},
			{
				Name:        "created_at",
				DataType:    generator.SQLDatetime,
				Constraints: []string{},
			},
			{
				Name:        "updated_at",
				DataType:    generator.SQLDatetime,
				Constraints: []string{},
			},
		},
		UniqueIndexes: []generator.UniqueIndex{},
		ForeignKeys: []generator.ForeignKey{
			{
				Table:      "tenants",
				FromColumn: "tenant_id",
				ToColumn:   "id",
			},
		},
	}
}

func TestCreateGen(t *testing.T) {
	createTemplate, createTestTemplate, err := generator.RenderCreateTemplate("github.com/gurch101/gowebutils", getTestUserSchema())
	if err != nil {
		t.Fatal(err)
	}

	testutils.AssertFileEqualsString(t, "snapshots/create_user.txt", string(createTemplate))
	testutils.AssertFileEqualsString(t, "snapshots/create_user_test.txt", string(createTestTemplate))
}

func TestCreateGenNoUniqueIndexNoConstraints(t *testing.T) {
	table := generator.Table{
		Name: "users",
		Fields: []generator.Field{
			{
				Name:        "id",
				DataType:    generator.SQLInt64,
				Constraints: []string{},
			},
			{
				Name:        "version",
				DataType:    generator.SQLInt64,
				Constraints: []string{},
			},
			{
				Name:        "name",
				DataType:    generator.SQLString,
				Constraints: []string{},
			},
			{
				Name:        "created_at",
				DataType:    generator.SQLDatetime,
				Constraints: []string{},
			},
			{
				Name:        "updated_at",
				DataType:    generator.SQLDatetime,
				Constraints: []string{},
			},
		},
		UniqueIndexes: []generator.UniqueIndex{},
	}

	createTemplate, createTestTemplate, err := generator.RenderCreateTemplate("github.com/gurch101/gowebutils", table)
	if err != nil {
		t.Fatal(err)
	}

	testutils.AssertFileEqualsString(t, "snapshots/create_user_no_unique_index_no_constraints.txt", string(createTemplate))
	testutils.AssertFileEqualsString(t, "snapshots/create_user_no_unique_index_no_constraints_test.txt", string(createTestTemplate))
}
