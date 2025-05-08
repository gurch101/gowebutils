package generator_test

import (
	"testing"

	"github.com/gurch101/gowebutils/pkg/generator"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestHelperGen(t *testing.T) {
	testHelperTemplate, err := generator.RenderTestHelperTemplate("github.com/gurch101/gowebutils", getTestUserSchema())
	if err != nil {
		t.Fatal(err)
	}

	testutils.AssertFileEqualsString(t, "snapshots/test_helpers.txt", string(testHelperTemplate))
}

func TestHelperGen_ImmutableModel(t *testing.T) {
	immutableTable := generator.Table{
		Name: "users",
		Fields: []generator.Field{
			{
				Name:        "id",
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
	testHelperTemplate, err := generator.RenderTestHelperTemplate("github.com/gurch101/gowebutils", immutableTable)

	if err != nil {
		t.Fatal(err)
	}

	testutils.AssertFileEqualsString(t, "snapshots/test_helpers_immutable.txt", string(testHelperTemplate))
}
