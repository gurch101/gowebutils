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
