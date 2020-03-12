package stringutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type assertStrings func(*testing.T, []string)

var assertNil = func(t *testing.T, es []string) { assert.Nil(t, es) }

func TestSet(t *testing.T) {
	for _, tCase := range []struct {
		in     [][]string
		assert assertStrings
	}{
		{
			in:     [][]string{{"foo", "bar", "foo"}},
			assert: assertElementsMatch("foo", "bar"),
		},
		{
			in:     [][]string{{"foo"}, {"bar"}, {"foo"}},
			assert: assertElementsMatch("foo", "bar"),
		},
		{
			in:     [][]string{},
			assert: assertNil,
		},
		{
			in:     [][]string{{}},
			assert: assertNil,
		},
	} {
		var s Set

		for _, b := range tCase.in {
			s.Add(b...)
		}

		tCase.assert(t, s.Strings())
	}
}

func assertElementsMatch(ss ...string) assertStrings {
	return func(t *testing.T, es []string) { assert.ElementsMatch(t, es, ss) }
}
