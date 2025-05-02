package stringutils

import (
	"strings"
	"unicode"
)

// SnakeToHuman converts snake_case to Snake Case.
func SnakeToHuman(s string) string {
	words := strings.Split(s, "_")
	for i := range words {
		if len(words[i]) > 0 {
			// Capitalize first letter and lowercase the rest
			words[i] = string(unicode.ToUpper(rune(words[i][0]))) + words[i][1:]
		}
	}

	return strings.Trim(strings.Join(words, " "), " ")
}
