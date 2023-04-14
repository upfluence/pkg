//go:build go1.18

package slices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReference(t *testing.T) {
	var (
		i = 1
		j = 2
	)

	assert.Equal(
		t,
		[]*int{&i, &j},
		References([]int{i, j}),
	)
}

func TestIndirectSlice(t *testing.T) {
	var (
		i = 1
		j = 2
	)

	assert.Equal(
		t,
		[]int{i, j},
		Indirect([]*int{&i, &j}),
	)
}
