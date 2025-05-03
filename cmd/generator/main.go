package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/generator"
	"github.com/gurch101/gowebutils/pkg/parser"
	_ "github.com/mattn/go-sqlite3"
)

const (
	generatedFilePermission      = 0o644
	generatedDirectoryPermission = 0o755
)

func main() {
	module, err := generator.GetModuleNameFromGoMod()
	if err != nil {
		panic(err)
	}

	db := dbutils.OpenDBPool(parser.ParseEnvStringPanic("DB_FILEPATH"))
	defer db.Close()

	tableSchema, err := generator.ParseSchema(db)
	if err != nil {
		panic(err)
	}

	if _, err := os.Stat("internal"); os.IsNotExist(err) {
		if err := os.Mkdir("internal", generatedDirectoryPermission); err != nil {
			panic(err)
		}
	}

	runCli(module, tableSchema)
}

func runCli(module string, tableSchema []generator.Table) {
	selectedTables := []generator.Table{tableSchema[0]}
	selectedActions := []string{"create", "get", "delete", "update", "list"}

	type actionConfig struct {
		renderFunc  func(string, generator.Table) ([]byte, []byte, error)
		fileNameFmt string
		testFileFmt string
	}

	actionMap := map[string]actionConfig{
		"create": {
			generator.RenderCreateTemplate,
			"internal/%s/create_%s.go",
			"internal/%s/create_%s_test.go",
		},
		"get": {
			generator.RenderGetOneTemplate,
			"internal/%s/get_%s_by_id.go",
			"internal/%s/get_%s_by_id_test.go",
		},
		"update": {
			generator.RenderUpdateTemplate,
			"internal/%s/update_%s.go",
			"internal/%s/update_%s_test.go",
		},
		"list": {
			generator.RenderSearchTemplate,
			"internal/%s/search_%s.go",
			"internal/%s/search_%s_test.go",
		},
		"delete": {
			generator.RenderDeleteTemplate,
			"internal/%s/delete_%s_by_id.go",
			"internal/%s/delete_%s_by_id_test.go",
		},
	}

	for _, table := range selectedTables {
		singularModelName := strings.ToLower(strings.TrimSuffix(table.Name, "s"))

		for _, action := range selectedActions {
			cfg := actionMap[action]

			template, testTemplate, err := cfg.renderFunc(module, table)
			if err != nil {
				panic(err)
			}

			filename := fmt.Sprintf(cfg.fileNameFmt, table.Name, singularModelName)
			testFilename := fmt.Sprintf(cfg.testFileFmt, table.Name, singularModelName)

			writeFileIfNotExist(filename, template)
			writeFileIfNotExist(testFilename, testTemplate)
		}
	}
}

func writeFileIfNotExist(filename string, content []byte) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if err := os.WriteFile(filename, content, generatedFilePermission); err != nil {
			panic(fmt.Errorf("error writing file: %w", err))
		}
	} else {
		//nolint: forbidigo
		fmt.Printf("File %s already exists. Skipping.\n", filename)
	}
}
