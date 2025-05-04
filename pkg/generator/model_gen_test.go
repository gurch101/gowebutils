package generator_test

import (
	"testing"

	"github.com/gurch101/gowebutils/pkg/generator"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestModelGen(t *testing.T) {
	modelTemplate, err := generator.RenderModelTemplate("github.com/gurch101/gowebutils", getTestUserSchema())
	if err != nil {
		t.Fatal(err)
	}

	testutils.AssertFileEqualsString(t, "snapshots/user_model.txt", string(modelTemplate))
}
