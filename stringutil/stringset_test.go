package stringutil

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type assertStrings func(*testing.T, []string)

func TestStringSet(t *testing.T) {
	for _, tCase := range []struct{
		in [][]string
		assert assertStrings
	} {
		{
			in: [][]string{
				{"foo", "bar", "foo"},
			},
			assert: assertElementsMatch("foo", "bar"),
		},
		{
			in: [][]string{
				{"foo"}, {"bar"}, {"foo"},
			},
			assert: assertElementsMatch("foo", "bar"),
		},
		{
			in: [][]string{},
			assert: assertNil(),
		},
	} {
		ss := StringSet{}

		for _, bs := range tCase.in {
			ss.Add(bs...)
		}

		tCase.assert(t, ss.Strings())
	}
}


func assertElementsMatch(ss ...string) assertStrings {
	return func(t *testing.T, es []string) {
		assert.ElementsMatch(t, es, ss)
	}
}


func assertNil() assertStrings {
	return func(t *testing.T, es []string) {
		assert.Nil(t, es)
	}
}
