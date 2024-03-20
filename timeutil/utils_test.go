package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnixOrNil(t *testing.T) {
	assert.Nil(t, UnixOrNil(time.Time{}))
	assert.Equal(t, int64(123456), *UnixOrNil(time.Unix(123456, 0)))
}

func TestUnixOrZero(t *testing.T) {
	assert.Equal(t, int64(0), UnixOrZero(time.Time{}))
	assert.Equal(t, int64(123456), UnixOrZero(time.Unix(123456, 0)))
}
