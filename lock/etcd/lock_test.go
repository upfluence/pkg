package etcd

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/upfluence/pkg/lock"
	"github.com/upfluence/pkg/lock/locktest"
	"github.com/upfluence/pkg/testutil"
)

func TestIntegration(t *testing.T) {
	url := testutil.FetchEnvVariable(t, "ETCD_URL")

	locktest.IntegrationTest(t, func(t testing.TB) lock.LockManager {
		lm, err := NewLockManager(url, "lock-testing", "/pkg/testing")

		require.Nil(t, err)
		return lm
	})
}
