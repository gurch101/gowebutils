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
	selectedTables := getTableSelection(tableSchema)
	selectedActions := getActionSelection()

	actionMap := getActionMap()

	for _, table := range selectedTables {
		singularModelName := strings.ToLower(strings.TrimSuffix(table.Name, "s"))

		if _, err := os.Stat("internal/" + table.Name); os.IsNotExist(err) {
			if err := os.Mkdir("internal/"+table.Name, generatedDirectoryPermission); err != nil {
				panic(err)
			}
		}

		for _, action := range selectedActions {
			if action == "model" {
				modelTemplate, err := generator.RenderModelTemplate(module, table)
				if err != nil {
					panic(err)
				}

				writeFileIfNotExist(fmt.Sprintf("internal/%s/models.go", table.Name), modelTemplate)

				continue
			}

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

func getActionSelection() []string {
	allActions := []string{"create", "get", "update", "list", "delete", "model"}

	printActions(allActions)

	for {
		actionNames := getUserInput("Enter actions (comma-separated): ")

		normalized := normalizeInput(actionNames)
		if isSelectAll(normalized) {
			return allActions
		}

		selectedActions, invalid := filterValidActions(normalized, allActions)
		if invalid {
			fmt.Println("One or more actions are invalid.")
			continue
		}

		return selectedActions
	}
}

func printActions(actions []string) {
	fmt.Println("Select actions to generate:")

	for _, action := range actions {
		fmt.Println(action)
	}
}

func filterValidActions(actions, validActions []string) ([]string, bool) {
	var result []string

	for _, action := range actions {
		if !contains(validActions, action) {
			fmt.Printf("Action %s does not exist.\n", action)
			return nil, true
		}

		result = append(result, action)
	}

	return result, false
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}

	return false
}

type actionConfig struct {
	renderFunc  func(string, generator.Table) ([]byte, []byte, error)
	fileNameFmt string
	testFileFmt string
}

func getActionMap() map[string]actionConfig {
	return map[string]actionConfig{
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
}
