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

func assertElementsMatch(ss ...int64) assertInt64s {
	return func(t *testing.T, es []int64) { assert.ElementsMatch(t, es, ss) }
}
