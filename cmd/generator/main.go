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
	var selectedTables []generator.Table

	// print table names
	fmt.Println("Select tables to generate:")
	for _, table := range tableSchema {
		fmt.Printf("%s\n", table.Name)
	}

	// prompt user to write table
	var tableNames []string

	for {
		fmt.Print("Enter table names (comma-separated): ")
		var input string
		if _, err := fmt.Scanln(&input); err != nil {
			panic(err)
		}
		tableNames = strings.Split(input, ",")
		if len(tableNames) == 0 {
			fmt.Println("Please enter at least one table name.")
			continue
		}

		for i, tableName := range tableNames {
			tableNames[i] = strings.ToLower(strings.TrimSpace(tableName))
		}

		if tableNames[0] == "all" || tableNames[0] == "*" {
			selectedTables = tableSchema
			break
		}

		invalidTable := false
		// ensure every table name is valid
		for _, tableName := range tableNames {
			found := false
			for _, table := range tableSchema {
				if table.Name == tableName {
					found = true
					selectedTables = append(selectedTables, table)
					break
				}
			}
			if !found {
				fmt.Printf("Table %s does not exist.\n", tableName)
				invalidTable = true
				break
			}
		}

		if invalidTable {
			continue
		}

		break
	}

	return selectedTables
}

func getActionSelection() []string {
	var selectedActions []string

	allActions := []string{"create", "get", "update", "list", "delete", "model"}

	for {
		fmt.Println("Select actions to generate:")
		for _, action := range allActions {
			fmt.Printf("%s\n", action)
		}
		fmt.Print("Enter actions (comma-separated): ")
		var input string
		if _, err := fmt.Scanln(&input); err != nil {
			panic(err)
		}
		selectedActions = strings.Split(input, ",")
		if len(selectedActions) == 0 {
			fmt.Println("Please enter at least one action.")
			continue
		}

		for i, action := range selectedActions {
			selectedActions[i] = strings.ToLower(strings.TrimSpace(action))
		}

		if selectedActions[0] == "all" || selectedActions[0] == "*" {
			selectedActions = allActions
			break
		}

		// ensure every action is valid
		invalidAction := false
		for _, action := range selectedActions {
			found := false
			for _, validAction := range allActions {
				if action == validAction {
					found = true
					selectedActions = append(selectedActions, action)
					break
				}
			}

			if !found {
				fmt.Printf("Action %s does not exist.\n", action)
				invalidAction = true
				break
			}
		}

		if invalidAction {
			continue
		}
		break
	}

	return selectedActions
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
