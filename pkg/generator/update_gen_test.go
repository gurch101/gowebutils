package generator_test

import (
	"testing"

	"github.com/gurch101/gowebutils/pkg/generator"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestUpdateGen(t *testing.T) {
	updateTemplate, updateTestTemplate, err := generator.RenderUpdateTemplate("github.com/gurch101/gowebutils", testUserSchema)
	if err != nil {
		t.Fatal(err)
	}

	testutils.AssertFileEqualsString(t, "snapshots/update_user.txt", string(updateTemplate))
	testutils.AssertFileEqualsString(t, "snapshots/update_user_test.txt", string(updateTestTemplate))
}

func TestUpdateGenNoConstraints(t *testing.T) {
	table := generator.Table{
		Name: "users",
		Fields: []generator.Field{
			{
				Name:     "id",
				DataType: generator.SQLInt64,
			},
			{
				Name:     "version",
				DataType: generator.SQLInt64,
			},
			{
				Name:     "name",
				DataType: generator.SQLString,
			},
			{
				Name:     "created_at",
				DataType: generator.SQLDatetime,
			},
			{
				Name:     "updated_at",
				DataType: generator.SQLDatetime,
			},
		},
		UniqueIndexes: []generator.UniqueIndex{},
	}

	updateTemplate, updateTestTemplate, err := generator.RenderUpdateTemplate("github.com/gurch101/gowebutils", table)
	if err != nil {
		t.Fatal(err)
	}

	testutils.AssertFileEqualsString(t, "snapshots/update_user_no_constraints.txt", string(updateTemplate))
	testutils.AssertFileEqualsString(t, "snapshots/update_user_no_constraints_test.txt", string(updateTestTemplate))
}
