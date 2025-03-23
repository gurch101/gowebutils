// Package templateutils is a package that provides utility functions for loading and using templates.
package templateutils

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
)

// ErrParseTemplate is an error that occurs when template parsing fails.
var ErrParseTemplate = errors.New("failed to parse template")

// LoadTemplates loads all the templates from the given embed.FS and returns a map of templates.
// Panics if any error occurs.
func LoadTemplates(templateFS embed.FS) map[string]*template.Template {
	templates := make(map[string]*template.Template)

	err := fs.WalkDir(templateFS, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Check if it's a file (not a directory)
		if !entry.IsDir() {
			tmpl, err := template.ParseFS(templateFS, path)
			if err != nil {
				return ErrParseTemplate
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
