package pointers

import (
	"testing"
	"time"

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

type testStruct struct {
	number int
	string string
}

func TestEq(t *testing.T) {
	n := 5
	assert.True(t, Equal(&n, &n))
	assert.True(t, Equal(Ptr(5), &n))
	assert.True(t, Equal(Ptr(5), Ptr(5)))
	assert.False(t, Equal(&n, Ptr(6)))
	assert.False(t, Equal(Ptr(5), Ptr(6)))

	str := "foo"
	assert.True(t, Equal(&str, &str))
	assert.True(t, Equal(&str, Ptr("foo")))
	assert.False(t, Equal(&str, Ptr("bar")))

	s := testStruct{
		number: 1,
		string: "hello",
	}
	assert.True(t, Equal(&s, &s))
	assert.True(t, Equal(&s, &testStruct{
		number: 1,
		string: "hello",
	}))
	assert.False(t, Equal(&s, &testStruct{
		number: 2,
		string: "hello",
	}))
	assert.False(t, Equal(&s, &testStruct{
		number: 1,
		string: "world",
	}))

	t0 := time.Now()
	assert.True(t, Equal(Ptr(t0), Ptr(t0)))
	assert.False(t, Equal(Ptr(t0), Ptr(t0.Add(time.Hour))))
	assert.True(t, Equal(Ptr(t0.Add(time.Hour)), Ptr(t0.Add(time.Hour))))
}
