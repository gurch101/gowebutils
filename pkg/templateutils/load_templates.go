// Package templateutils is a package that provides utility functions for loading and using templates.
package templateutils

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"text/template"
)

var errReadTemplate = errors.New("failed to read template")

// LoadTemplates loads all the templates from the given embed.FS and returns a map of templates.
// Panics if any error occurs.
func LoadTemplates(templateFS embed.FS, path string) map[string]*template.Template {
	templates := make(map[string]*template.Template)

	err := fs.WalkDir(templateFS, path, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Check if it's a file (not a directory)
		if !entry.IsDir() {
			tmpl, err := template.ParseFS(templateFS, path)
			if err != nil {
				return errReadTemplate
			}
			// Store file content in the map using the file name as the key
			templates[entry.Name()] = tmpl
		}

		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("error loading templates: %v", err))
	}

	return templates
}
