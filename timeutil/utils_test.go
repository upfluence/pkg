package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnixOrNil(t *testing.T) {
	var n *int64

	assert.Equal(t, UnixOrNil(time.Time{}), n)
	assert.Equal(t, *UnixOrNil(time.Unix(123456, 0)), time.Unix(123456, 0).Unix())
}
