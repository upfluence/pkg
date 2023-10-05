package pointers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPtr(t *testing.T) {
	assert.Equal(
		t,
		"test",
		*(Ptr("test")),
	)
}

func TestNullablePtr(t *testing.T) {
	assert.Equal(t, -5, *NullablePtr(-5))
	assert.Nil(t, NullablePtr(0))

	assert.Equal(t, *NullablePtr("string"), "string")
	assert.Nil(t, NullablePtr(""))
}

func TestNullIsZero(t *testing.T) {
	var (
		i *int
		s *string
	)

	assert.Equal(t, 1, NullIsZero(Ptr(1)))
	assert.Equal(t, 0, NullIsZero(i))

	assert.Equal(t, "s", NullIsZero(Ptr("s")))
	assert.Equal(t, "", NullIsZero(s))
}
