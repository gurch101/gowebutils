package stringutils

import (
	"strings"
	"unicode"
)

func CamelToSnake(s string) string {
	var result strings.Builder

	runes := []rune(s)
	length := len(runes)

	i := 0
	for i < length {
		// Detect "ID" and preserve it
		if i+1 < length && runes[i] == 'I' && runes[i+1] == 'D' {
			if i > 0 {
				result.WriteRune('_')
			}

			result.WriteString("id")

			i += 2

			continue
		}

		r := runes[i]
		if unicode.IsUpper(r) && i > 0 && runes[i-1] != '_' {
			result.WriteRune('_')
		}

		result.WriteRune(unicode.ToLower(r))

		i++
	}

	return result.String()
}
