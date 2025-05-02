package stringutils

import (
	"strings"
)

// SnakeToTitle converts snake_case to TitleCase.
func SnakeToTitle(s string) string {
	words := strings.Split(s, "_")
	for i := range words {
		if len(words[i]) > 0 {
			if strings.ToLower(words[i]) != "id" {
				words[i] = strings.ToUpper(string(words[i][0])) + strings.ToLower(words[i][1:])
			} else {
				words[i] = "ID"
			}
		}
	}

	return strings.Join(words, "")
}
