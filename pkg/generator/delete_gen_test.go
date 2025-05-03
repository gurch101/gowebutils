package generator_test

import (
	"testing"

	"github.com/gurch101/gowebutils/pkg/generator"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestDeleteGen(t *testing.T) {
	deleteTemplate, deleteTestTemplate, err := generator.RenderDeleteTemplate("github.com/gurch101/gowebutils", testUserSchema)
	if err != nil {
		t.Fatal(err)
	}

	testutils.AssertFileEqualsString(t, "snapshots/delete_user_by_id.txt", string(deleteTemplate))
	testutils.AssertFileEqualsString(t, "snapshots/delete_user_by_id_test.txt", string(deleteTestTemplate))
}
