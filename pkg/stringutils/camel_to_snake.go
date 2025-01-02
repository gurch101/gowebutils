package stringutils

import (
	"strings"
	"unicode"
)

func CamelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			// If the character is uppercase and not the first character, add an underscore
			result.WriteRune('_')
		}
		// Convert the character to lowercase and add it to the result
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}
