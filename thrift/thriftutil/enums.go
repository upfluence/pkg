package thriftutil

import (
	"strings"

	"github.com/upfluence/pkg/stringutil"
)

func SanitizeCamelCaseThriftEnumValue(v string) string {
	if sks := strings.SplitN(v, "_", 2); len(sks) == 2 {
		return stringutil.CamelToSnakeCase(sks[1])
	}

	return strings.ToLower(v)
}
