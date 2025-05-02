package stringutils

import "strings"

// SnakeToKebab converts snake_case to kebab-case (replaces underscores with dashes).
func SnakeToKebab(s string) string {
	return strings.ReplaceAll(s, "_", "-")
}
