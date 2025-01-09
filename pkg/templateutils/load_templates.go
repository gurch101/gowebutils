package templateutils

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"text/template"
)

var ErrReadTemplate = errors.New("failed to read template")

var ErrReadTemplateFS = errors.New("failed to read template fs")

func LoadTemplates(templateFS embed.FS, path string) (map[string]*template.Template, error) {
	templates := make(map[string]*template.Template)

	err := fs.WalkDir(templateFS, path, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Check if it's a file (not a directory)
		if !entry.IsDir() {
			tmpl, err := template.ParseFS(templateFS, path)
			if err != nil {
				return ErrReadTemplate
			}
			// Store file content in the map using the file name as the key
			templates[entry.Name()] = tmpl
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadTemplateFS, err)
	}

	return templates, nil
}
