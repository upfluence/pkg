//go:build go1.18

package generics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPointer(t *testing.T) {
	assert.Equal(
		t,
		"test",
		*(Pointer("test")),
	)
}

func TestNullablePtr(t *testing.T) {
	assert.Equal(t, -5, *NullablePtr(-5))
	assert.Nil(t, NullablePtr(0))

	assert.Equal(t, *NullablePtr("string"), "string")
	assert.Nil(t, NullablePtr(""))
}

func TestReferenceSlice(t *testing.T) {
	var (
		i = 1
		j = 2
	)

	assert.Equal(
		t,
		[]*int{&i, &j},
		ReferenceSlice([]int{i, j}),
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
		IndirectSlice([]*int{&i, &j}),
	)
}
