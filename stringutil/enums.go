package stringutil

import (
	"regexp"
	"strings"
)

var upperCase = regexp.MustCompile(`([A-Z]+)`)

func SanitizeCamelCaseThriftEnumValue(v string) string {
	if sks := strings.SplitN(v, "_", 2); len(sks) == 2 {
		v = upperCase.ReplaceAllString(sks[1], `_$1`)
		v = strings.TrimPrefix(v, "_")
	}

	return strings.ToLower(v)
}
