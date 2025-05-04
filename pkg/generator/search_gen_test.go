package generator_test

import (
	"testing"

	"github.com/gurch101/gowebutils/pkg/generator"
	"github.com/gurch101/gowebutils/pkg/testutils"
)

func TestSearchGen(t *testing.T) {
	searchTemplate, searchTestTemplate, err := generator.RenderSearchTemplate("github.com/gurch101/gowebutils", getTestUserSchema())
	if err != nil {
		t.Fatal(err)
	}

	testutils.AssertFileEqualsString(t, "snapshots/search_user.txt", string(searchTemplate))
	testutils.AssertFileEqualsString(t, "snapshots/search_user_test.txt", string(searchTestTemplate))
}
