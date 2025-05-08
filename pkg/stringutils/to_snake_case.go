package stringutils

import (
	"regexp"
	"strings"
	"unicode"
)

// ToSnakeCase converts a string to snake_case.
func ToSnakeCase(str string) string {
	// Trim whitespace
	str = strings.TrimSpace(str)

	// Replace hyphens and spaces with underscores
	str = strings.ReplaceAll(str, "-", "_")
	str = strings.ReplaceAll(str, " ", "_")

	// Insert underscores before camel case boundaries (e.g., "HelloWorld" -> "Hello_World")
	var result []rune

	for i, r := range str {
		if i > 0 && unicode.IsUpper(r) && (unicode.IsLower(rune(str[i-1])) || (i+1 < len(str) && unicode.IsLower(rune(str[i+1])))) {
			result = append(result, '_')
		}

		result = append(result, unicode.ToLower(r))
	}

	// Remove multiple underscores
	re := regexp.MustCompile(`_+`)
	snake := re.ReplaceAllString(string(result), "_")

	return snake
}
