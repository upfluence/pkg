package slices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnique(t *testing.T) {
	var s = []int64{1, 3, 23}

	assert.Equal(
		t,
		map[int64]struct{}{1: {}, 3: {}, 23: {}},
		Unique(s),
	)
}

func TestUniquePtr(t *testing.T) {
	var s = References([]int64{1, 3, 23})

	assert.Equal(
		t,
		map[int64]struct{}{1: {}, 3: {}, 23: {}},
		UniquePtr(s),
	)
}
