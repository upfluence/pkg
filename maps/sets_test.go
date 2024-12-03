package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeys(t *testing.T) {
	var m = map[string]int{
		"foo":  2,
		"fizz": 1,
	}

	assert.ElementsMatch(t, []string{"foo", "fizz"}, Keys(m))
}

func TestValues(t *testing.T) {
	var m = map[string]int{
		"foo":  2,
		"fizz": 1,
	}

	assert.ElementsMatch(t, []int{1, 2}, Values(m))
}

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

type fooBar struct {
	key   string
	value int
}

func TestIndexedSet_Add(t *testing.T) {
	var elems = []fooBar{
		{"foo", 2},
		{"bar", 1},
		{"foo", 4},
	}
	for _, tt := range []struct {
		name       string
		in         [][]fooBar
		wantKeys   []string
		wantValues []fooBar
	}{
		{
			name: "single slice with duplicate",
			in: [][]fooBar{
				{elems[0], elems[1], elems[2]},
			},
			wantKeys:   []string{elems[2].key, elems[1].key},
			wantValues: []fooBar{elems[2], elems[1]},
		},
		{
			name: "multiple slices with duplicate",
			in: [][]fooBar{
				{elems[0]},
				{elems[1]},
				{elems[2]},
			},
			wantKeys:   []string{elems[2].key, elems[1].key},
			wantValues: []fooBar{elems[2], elems[1]},
		},
		{name: "empty slices", in: [][]fooBar{{}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var s = IndexedSet[string, fooBar]{
				Fn: func(t fooBar) string {
					return t.key
				},
			}

			for _, b := range tt.in {
				s.Add(b...)
			}

			assert.ElementsMatch(t, tt.wantValues, s.Values())
			assert.ElementsMatch(t, tt.wantKeys, s.Keys())
		})

	}
}

func TestIndexedSet_Has(t *testing.T) {
	var s Set[string]

	assert.False(t, s.Has("foo"))

	s.Add("foo")

	assert.True(t, s.Has("foo"))
	assert.False(t, s.Has("bar"))
}

func TestIndexedSet_Delete(t *testing.T) {
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
