package executor

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/pkg/v2/lock/local"
)

func TestExecute(t *testing.T) {
	var (
		i  int64
		lm local.LockManager

		err = NewExecutor(
			lm.Lock("foo"),
			func(context.Context) error {
				atomic.AddInt64(&i, 1)
				return nil
			},
		).Execute(context.Background())
	)

	assert.Nil(t, err)
	assert.Equal(t, int64(1), i)
}
