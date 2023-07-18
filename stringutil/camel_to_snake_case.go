package stringutil

import (
	"regexp"
	"strings"
)

var upperCase = regexp.MustCompile(`([A-Z]+)`)

func CamelToSnakeCase(s string) string {
	s = upperCase.ReplaceAllString(s, `_$1`)
	s = strings.TrimPrefix(s, "_")

	return strings.ToLower(s)
}
