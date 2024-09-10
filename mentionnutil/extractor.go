package mentionnutil

import (
	"regexp"
	"strings"

	"github.com/upfluence/pkg/slices"
	"github.com/upfluence/pkg/stringutil"
)

var mentionRegex = regexp.MustCompile(`\B@[\w\.]+[\w]+`)

func ExtractMentions(s string) []string {
	mentions := slices.Map(
		mentionRegex.FindAllString(s, -1),
		func(v string) string {
			return strings.TrimPrefix(v, "@")
		},
	)

	if len(mentions) == 0 {
		return nil
	}

	var set stringutil.Set

	set.Add(mentions...)

	return set.Strings()
}
