package int64util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type assertInt64s func(*testing.T, []int64)

var assertNil = func(t *testing.T, es []int64) { assert.Nil(t, es) }

func TestSet(t *testing.T) {

	for _, tt := range []struct {
		in     [][]int64
		assert assertInt64s
	}{
		{
			in:     [][]int64{{5, 6, 6}},
			assert: assertElementsMatch(5, 6),
		},
		{
			in:     [][]int64{{6}, {6}, {5}},
			assert: assertElementsMatch(5, 6),
		},
		{
			in:     [][]int64{},
			assert: assertNil,
		},
		{
			in:     [][]int64{{}},
			assert: assertNil,
		},
	} {
		var s Set

		for _, b := range tt.in {
			s.Add(b...)
		}

		tt.assert(t, s.Int64s())
	}
}

func TestHas(t *testing.T) {
	var s Set

	assert.False(t, s.Has(3))

	s.Add(3)

	assert.True(t, s.Has(3))
	assert.False(t, s.Has(8))
}

func TestDelete(t *testing.T) {
	var s Set

	s.Add(3)
	s.Add(4)
	s.Add(5)
	assert.True(t, s.Has(3))
	assert.True(t, s.Has(4))
	assert.True(t, s.Has(5))

	s.Delete(4, 5)
	assert.True(t, s.Has(3))
	assert.False(t, s.Has(4))
	assert.False(t, s.Has(5))
}

func assertElementsMatch(ss ...int64) assertInt64s {
	return func(t *testing.T, es []int64) { assert.ElementsMatch(t, es, ss) }
}
