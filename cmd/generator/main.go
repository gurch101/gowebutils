package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gurch101/gowebutils/pkg/collectionutils"
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
	selectedTables := getTableSelection(tableSchema)
	selectedActions := []string{"create", "get", "update", "list", "delete", "exists", "model", "routes", "test_helper"}

	actionMap := getActionMap()

	writeFileIfNotExist("internal/schema_test.go", []byte(generator.GetSchemaTest()))

	for _, table := range selectedTables {
		if _, err := os.Stat("internal/" + table.Name); os.IsNotExist(err) {
			if err := os.Mkdir("internal/"+table.Name, generatedDirectoryPermission); err != nil {
				panic(err)
			}
		}

		for _, action := range selectedActions {
			if action == "update" {
				ok := collectionutils.Contains(table.Fields, func(field generator.Field) bool {
					return field.Name == "version"
				})

				if !ok {
					fmt.Printf("Table %s does not have a version field. Skipping update action.\n", table.Name)
					continue
				}
			}

			cfg := actionMap[action]

			template, testTemplate, err := cfg.renderFunc(module, table)
			if err != nil {
				panic(err)
			}

			filename, testFilename := cfg.fileNameFunc(table)
			writeFileIfNotExist(filename, template)

			if testTemplate != nil {
				writeFileIfNotExist(testFilename, testTemplate)
			}
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

func getTableSelection(tableSchema []generator.Table) []generator.Table {
	printTableOptions(tableSchema)

	for {
		tableNames := getUserInput("Enter table names (comma-separated): ")

		normalized := normalizeInput(tableNames)
		if isSelectAll(normalized) {
			return tableSchema
		}

		selectedTables, invalid := filterValidTables(normalized, tableSchema)
		if invalid {
			fmt.Println("One or more table names are invalid.")
			continue
		}

		return selectedTables
	}
}

func printTableOptions(tables []generator.Table) {
	fmt.Println("Select tables to generate:")

	for _, table := range tables {
		fmt.Println(table.Name)
	}
}

func getUserInput(prompt string) []string {
	fmt.Print(prompt)

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		panic(err)
	}

	return strings.Split(input, ",")
}

func normalizeInput(inputs []string) []string {
	for i := range inputs {
		inputs[i] = strings.ToLower(strings.TrimSpace(inputs[i]))
	}

	return inputs
}

func isSelectAll(input []string) bool {
	if len(input) == 0 {
		return false
	}

	return input[0] == "all" || input[0] == "*"
}

func filterValidTables(names []string, schema []generator.Table) ([]generator.Table, bool) {
	var result []generator.Table

	for _, name := range names {
		found := false

		for _, table := range schema {
			if table.Name == name {
				result = append(result, table)
				found = true

				break
			}
		}

		if !found {
			fmt.Printf("Table %s does not exist.\n", name)
			return nil, true
		}
	}

	return result, false
}

type actionConfig struct {
	renderFunc   func(string, generator.Table) ([]byte, []byte, error)
	fileNameFunc func(generator.Table) (string, string)
}

// nolint: funlen
func getActionMap() map[string]actionConfig {
	return map[string]actionConfig{
		"create": {
			generator.RenderCreateTemplate,
			func(table generator.Table) (string, string) {
				singularModelName := strings.ToLower(strings.TrimSuffix(table.Name, "s"))
				return fmt.Sprintf("internal/%s/create_%s.go", strings.ToLower(table.Name), singularModelName),
					fmt.Sprintf("internal/%s/create_%s_test.go", strings.ToLower(table.Name), singularModelName)
			},
		},
		"get": {
			generator.RenderGetOneTemplate,
			func(table generator.Table) (string, string) {
				singularModelName := strings.ToLower(strings.TrimSuffix(table.Name, "s"))
				return fmt.Sprintf("internal/%s/get_%s_by_id.go", strings.ToLower(table.Name), singularModelName),
					fmt.Sprintf("internal/%s/get_%s_by_id_test.go", strings.ToLower(table.Name), singularModelName)
			},
		},
		"update": {
			generator.RenderUpdateTemplate,
			func(table generator.Table) (string, string) {
				singularModelName := strings.ToLower(strings.TrimSuffix(table.Name, "s"))
				return fmt.Sprintf("internal/%s/update_%s.go", strings.ToLower(table.Name), singularModelName),
					fmt.Sprintf("internal/%s/update_%s_test.go", strings.ToLower(table.Name), singularModelName)
			},
		},
		"list": {
			generator.RenderSearchTemplate,
			func(table generator.Table) (string, string) {
				return fmt.Sprintf("internal/%s/search_%s.go", strings.ToLower(table.Name), strings.ToLower(table.Name)),
					fmt.Sprintf("internal/%s/search_%s_test.go", strings.ToLower(table.Name), strings.ToLower(table.Name))
			},
		},
		"delete": {
			generator.RenderDeleteTemplate,
			func(table generator.Table) (string, string) {
				singularModelName := strings.ToLower(strings.TrimSuffix(table.Name, "s"))
				return fmt.Sprintf("internal/%s/delete_%s_by_id.go", strings.ToLower(table.Name), singularModelName),
					fmt.Sprintf("internal/%s/delete_%s_by_id_test.go", strings.ToLower(table.Name), singularModelName)
			},
		},
		"model": {
			func(module string, table generator.Table) ([]byte, []byte, error) {
				modelTemplate, err := generator.RenderModelTemplate(module, table)
				return modelTemplate, nil, err
			},
			func(table generator.Table) (string, string) {
				return fmt.Sprintf("internal/%s/models.go", strings.ToLower(table.Name)), ""
			},
		},
		"routes": {
			func(module string, table generator.Table) ([]byte, []byte, error) {
				routesTemplate, err := generator.RenderRoutesTemplate(module, table)
				return routesTemplate, nil, err
			},
			func(table generator.Table) (string, string) {
				return fmt.Sprintf("internal/%s/routes.go", strings.ToLower(table.Name)), ""
			},
		},
		"test_helper": {
			func(module string, table generator.Table) ([]byte, []byte, error) {
				testHelperTemplate, err := generator.RenderTestHelperTemplate(module, table)
				return testHelperTemplate, nil, err
			},
			func(table generator.Table) (string, string) {
				return fmt.Sprintf("internal/%s/test_helpers.go", strings.ToLower(table.Name)), ""
			},
		},
		"exists": {
			func(module string, table generator.Table) ([]byte, []byte, error) {
				existsTemplate, err := generator.RenderExistsTemplate(module, table)
				return existsTemplate, nil, err
			},
			func(table generator.Table) (string, string) {
				singularModelName := strings.ToLower(strings.TrimSuffix(table.Name, "s"))
				return fmt.Sprintf("internal/%s/%s_exists.go", strings.ToLower(table.Name), singularModelName), ""
			},
		},
	}
}
