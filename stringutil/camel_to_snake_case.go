package stringutil

import (
	"regexp"
	"strings"
)

var upperCaseRegexps = []*regexp.Regexp{
	regexp.MustCompile(`([A-Z][a-z]+)`),
	regexp.MustCompile(`([A-Z]+)`),
}

func CamelToSnakeCase(s string) string {
	for _, re := range upperCaseRegexps {
		s = re.ReplaceAllStringFunc(s, func(s string) string { return "_" + strings.ToLower(s) })
	}

	return strings.ToLower(strings.TrimPrefix(s, "_"))
}
