package thriftutil

import (
	"strings"

	"github.com/upfluence/pkg/v2/stringutil"
)

func SanitizeCamelCaseThriftEnumValue(v string) string {
	if sks := strings.SplitN(v, "_", 2); len(sks) == 2 {
		return strings.ReplaceAll(stringutil.CamelToSnakeCase(sks[1]), "__", "_")
	}

	return strings.ToLower(v)
}
