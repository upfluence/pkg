package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet_Add(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   [][]string
		want []string
	}{
		{
			name: "single slice with duplicate",
			in:   [][]string{{"foo", "bar", "foo"}},
			want: []string{"foo", "bar"},
		},
		{
			name: "multiple slices with duplicate",
			in:   [][]string{{"foo"}, {"bar"}, {"foo"}},
			want: []string{"foo", "bar"},
		},
		{name: "empty slices", in: [][]string{{}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var s Set[string]

			for _, b := range tt.in {
				s.Add(b...)
			}

			assert.ElementsMatch(t, tt.want, s.Keys())
		})
	}
}

func TestSet_Has(t *testing.T) {
	var s Set[string]

	assert.False(t, s.Has("foo"))

	s.Add("foo")

	assert.True(t, s.Has("foo"))
	assert.False(t, s.Has("bar"))
}

func TestSet_Delete(t *testing.T) {
	var s Set[string]

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
