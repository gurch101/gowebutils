package dbutils

import "github.com/gurch101/gowebutils/pkg/stringutils"

func BuildSearchSelectFields(tableName string, fields []string, customMappings map[string]string) []string {
	dbFields := make([]string, len(fields)+1)
	dbFields[0] = "count(*) over()"

	for i, field := range fields {
		// Check if there's a custom mapping for this field
		if mappedField, exists := customMappings[field]; exists {
			dbFields[i+1] = mappedField
		} else {
			dbFields[i+1] = tableName + "." + stringutils.CamelToSnake(field)
		}
	}

	return dbFields
}
