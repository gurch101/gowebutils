package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"text/template"
)

func renderTemplateFile(templateString string, data interface{}) ([]byte, error) {
	funcMap := template.FuncMap{
		"incr": func(i int) int { return i + 1 },
	}
	tmpl := template.New("handler").Funcs(funcMap)

	tmpl, err := tmpl.Parse(templateString)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("error executing template: %w", err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error formatting generated code: %w", err)
	}

	// Return the formatted code
	return formatted, nil
}
