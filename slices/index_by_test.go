package slices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type dummy struct {
	ID int64
}

func TestIndexBy(t *testing.T) {
	assert.Equal(
		t,
		IndexBy([]int64{1, 2, 3}, func(v int64) int64 { return v }),
		map[int64]int64{1: 1, 2: 2, 3: 3},
	)

	var (
		d1 = dummy{ID: 1}
		d2 = dummy{ID: 2}
		d3 = dummy{ID: 3}
	)

	assert.Equal(
		t,
		IndexBy([]dummy{d1, d2, d3}, func(v dummy) int64 { return v.ID }),
		map[int64]dummy{1: d1, 2: d2, 3: d3},
	)

	assert.Equal(
		t,
		IndexBy([]*dummy{&d1, &d2, &d3}, func(v *dummy) int64 { return v.ID }),
		map[int64]*dummy{1: &d1, 2: &d2, 3: &d3},
	)
}
