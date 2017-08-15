package mock

import (
	"testing"

	"github.com/upfluence/pkg/stringcache/testutil"
)

func TestIntegration(t *testing.T) {
	testutil.IntegrationScenario(t, NewCache())
}
