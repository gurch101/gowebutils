package generator_test

import (
	"testing"

	"github.com/gurch101/gowebutils/pkg/generator"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestGetGen(t *testing.T) {
	template, testTemplate, err := generator.RenderGetOneTemplate("github.com/gurch101/gowebutils", getTestUserSchema())
	if err != nil {
		t.Fatal(err)
	}

	testutils.AssertFileEqualsString(t, "snapshots/get_user_by_id.txt", string(template))
	testutils.AssertFileEqualsString(t, "snapshots/get_user_by_id_test.txt", string(testTemplate))
}
