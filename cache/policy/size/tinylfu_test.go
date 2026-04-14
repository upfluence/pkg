package size

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/upfluence/pkg/v2/cache/policy"
	"github.com/upfluence/pkg/v2/cache/policy/policytest"
)

func TestTinyLFUPolicyHarness(t *testing.T) {
	policytest.RunTests(t, func(_ testing.TB) policy.EvictionPolicy[string] {
		// Capacity 4: small enough to trigger evictions in Set-heavy workloads
		// while large enough for the duplicate-set and explicit-evict tests.
		return NewTinyLFUPolicy[string](4)
	})
}

func BenchmarkTinyLFUPolicy(b *testing.B) {
	policytest.RunBenchmarks(b, func(_ testing.TB) policy.EvictionPolicy[string] {
		return NewTinyLFUPolicy[string](32)
	})
}

// TestTinyLFUAdmissionFrequencyWins verifies the core W-TinyLFU invariant: a
// key that has been accessed many times is preferred over a brand-new key when
// both compete for the last slot in main.
//
// Setup (capacity=10, window≈1, main≈9):
//  1. Fill the cache with keys "a".."i" so main is full and "a" is the LRU.
//  2. Access "a" many times so its frequency far exceeds zero.
//  3. Insert a new key "z" — it displaces "a" from the window into main.
//  4. At that point "z" (freq≈0 in sketch after door-keeper) should lose to
//     "a" in the TinyLFU admission test, so "z" is evicted, not "a".
func TestTinyLFUAdmissionFrequencyWins(t *testing.T) {
	p := NewTinyLFUPolicy[string](10)
	defer p.Close()

	for _, k := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"} {
		require.NoError(t, p.Op(k, policy.Set))
	}

	for range 20 {
		require.NoError(t, p.Op("a", policy.Get))
	}

	drainTinyLFUPending(p.C())

	require.NoError(t, p.Op("z", policy.Set))

	evicted := drainTinyLFUPending(p.C())
	for _, k := range evicted {
		assert.NotEqual(t, "a", k, "high-frequency key 'a' must not be evicted")
	}
}

// TestTinyLFUExplicitEvict verifies that an explicit Evict removes the key
// without sending on C().
func TestTinyLFUExplicitEvict(t *testing.T) {
	p := NewTinyLFUPolicy[string](4)

	require.NoError(t, p.Op("foo", policy.Set))
	require.NoError(t, p.Op("foo", policy.Evict))

	select {
	case k, ok := <-p.C():
		if ok {
			t.Errorf("unexpected eviction event for key %q", k)
		}
	default:
	}

	require.NoError(t, p.Close())
}

// TestTinyLFUSketchSaturation exercises the sketch halving path.
func TestTinyLFUSketchSaturation(t *testing.T) {
	p := NewTinyLFUPolicy[string](4)
	defer p.Close()

	go func() {
		for range p.C() {
		}
	}()

	for range 500 {
		p.Op("hot", policy.Set) //nolint:errcheck
	}
	require.NoError(t, p.Op("hot", policy.Get))
}

// TestTinyLFUDoorkeeperAging exercises the door-keeper reset path.
func TestTinyLFUDoorkeeperAging(t *testing.T) {
	p := NewTinyLFUPolicy[string](4)
	defer p.Close()

	go func() {
		for range p.C() {
		}
	}()

	for i := range 500 {
		k := string(rune('A' + i%26))
		p.Op(k, policy.Set) //nolint:errcheck
	}

	require.NoError(t, p.Op("z", policy.Get))
}

func drainTinyLFUPending(ch <-chan string) []string {
	var out []string
	for {
		select {
		case k, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, k)
		default:
			return out
		}
	}
}
