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

func TestHas(t *testing.T) {
	var s Set

	assert.False(t, s.Has("foo"))

	s.Add("foo")

	assert.True(t, s.Has("foo"))
	assert.False(t, s.Has("bar"))
}

func TestDelete(t *testing.T) {
	var s Set

	assert.False(t, s.Has("foo"))

	s.Add("foo")
	s.Add("bar")
	s.Add("biz")
	assert.True(t, s.Has("foo"))
	assert.True(t, s.Has("bar"))
	assert.True(t, s.Has("biz"))

	s.Delete("bar", "biz")
	assert.True(t, s.Has("foo"))
	assert.False(t, s.Has("bar"))
	assert.False(t, s.Has("biz"))
}

func assertElementsMatch(ss ...string) assertStrings {
	return func(t *testing.T, es []string) { assert.ElementsMatch(t, es, ss) }
}
