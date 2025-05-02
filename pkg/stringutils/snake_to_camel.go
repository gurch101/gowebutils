package stringutils

import "strings"

// SnakeToCamel converts snake_case to camelCase.
func SnakeToCamel(s string) string {
	word := SnakeToTitle(s)

	if len(word) == 0 {
		return ""
	}

	// if the word is "ID", return "id"
	// if the word ends with "ID", return wordId
	if word == "ID" {
		return "id"
	}

	if strings.HasSuffix(word, "ID") {
		word = strings.TrimSuffix(word, "ID") + "Id"
	}

	return strings.ToLower(string(word[0])) + word[1:]
}
