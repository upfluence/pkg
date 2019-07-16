package local

import (
	"testing"

	"github.com/upfluence/pkg/lock"
	"github.com/upfluence/pkg/lock/locktest"
)

func TestIntegration(t *testing.T) {
	locktest.IntegrationTest(t, func(testing.TB) lock.LockManager {
		return &LockManager{}
	})
}
