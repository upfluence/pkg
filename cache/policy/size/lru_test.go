package size

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/upfluence/pkg/v2/cache/policy"
	"github.com/upfluence/pkg/v2/cache/policy/policytest"
)

func TestLRUPolicy(t *testing.T) {
	p := NewLRUPolicy[string](2)

	p.Op("foo", policy.Set)
	p.Op("bar", policy.Set)
	p.Op("buz", policy.Set)

	k := <-p.C()
	assert.Equal(t, "foo", k)

	p.Op("bar", policy.Get)
	p.Op("foo", policy.Set)

	k = <-p.C()
	assert.Equal(t, "buz", k)

	p.Close()

	assert.Equal(t, policy.ErrClosed, p.Op("foo", policy.Set))
}

func TestLRUPolicyHarness(t *testing.T) {
	policytest.RunTests(t, func(_ testing.TB) policy.EvictionPolicy[string] {
		// Size 4: small enough that Set-only workloads trigger evictions,
		// large enough that the explicit-evict and duplicate-set tests work.
		return NewLRUPolicy[string](4)
	})
}

func BenchmarkLRUPolicy(b *testing.B) {
	policytest.RunBenchmarks(b, func(_ testing.TB) policy.EvictionPolicy[string] {
		return NewLRUPolicy[string](32)
	})
}
