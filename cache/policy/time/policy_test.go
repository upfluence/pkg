package time

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/upfluence/pkg/v2/cache/policy"
	"github.com/upfluence/pkg/v2/cache/policy/policytest"
	"github.com/upfluence/pkg/v2/timeutil/timetest"
)

// testPolicy bundles a tracker with its own channel so targeted tests can
// call cleanup() directly and read evictions from ch.
type testPolicy[K comparable] struct {
	*tracker[K]
	ch <-chan K
}

func newIdleTracker[K comparable](ttl time.Duration) (testPolicy[K], *timetest.Clock) {
	fc := &timetest.Clock{}
	ch := make(chan K, 16)
	evict := func(k K) { ch <- k }
	t := newTracker[K](ttl, func(tr *tracker[K]) func(K) { return tr.move }, evict, fc)
	return testPolicy[K]{tracker: t, ch: ch}, fc
}

func newLifetimeTracker[K comparable](ttl time.Duration) (testPolicy[K], *timetest.Clock) {
	fc := &timetest.Clock{}
	ch := make(chan K, 16)
	evict := func(k K) { ch <- k }
	t := newTracker[K](ttl, func(*tracker[K]) func(K) { return func(K) {} }, evict, fc)
	return testPolicy[K]{tracker: t, ch: ch}, fc
}

// collectN reads n keys from ch, returning them in order.
func collectN(t *testing.T, ch <-chan string, n int) []string {
	t.Helper()
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		select {
		case k, ok := <-ch:
			if !ok {
				t.Fatalf("channel closed prematurely after %d/%d keys", i, n)
			}
			out = append(out, k)
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for eviction %d/%d", i+1, n)
		}
	}
	return out
}

func TestIdlePolicy(t *testing.T) {
	base, fc := newIdleTracker[string](time.Second)

	// t=0: insert foo, bar, buz; Get foo; Evict bar.
	base.Op("foo", policy.Set)
	base.Op("bar", policy.Set)
	base.Op("buz", policy.Set)
	base.Op("foo", policy.Get)
	base.Op("bar", policy.Evict)

	// Advance to t=1.5s — both buz and foo are idle for >1s.
	fc.MoveBy(1500 * time.Millisecond)

	go base.cleanup()

	keys := collectN(t, base.ch, 2)
	assert.Equal(t, []string{"buz", "foo"}, keys)

	assert.Nil(t, base.Close())
}

func TestIdlePolicyTimestampRefresh(t *testing.T) {
	base, fc := newIdleTracker[string](time.Second)

	base.Op("foo", policy.Set) // inserted at t=0

	// t=0.8s: Get foo — refreshes its idle timer to t=0.8s.
	fc.MoveBy(800 * time.Millisecond)
	base.Op("foo", policy.Get)

	// t=1.2s: foo idle for only 0.4s — must NOT be evicted.
	fc.MoveBy(400 * time.Millisecond)
	base.cleanup()

	select {
	case k := <-base.ch:
		t.Errorf("unexpected eviction of %q at t=1.2s (idle only 0.4s)", k)
	default:
	}

	// t=1.9s: foo idle for 1.1s > 1s — must be evicted.
	fc.MoveBy(700 * time.Millisecond)
	go base.cleanup()

	keys := collectN(t, base.ch, 1)
	assert.Equal(t, []string{"foo"}, keys)

	assert.Nil(t, base.Close())
}

func TestLifetimePolicy(t *testing.T) {
	base, fc := newLifetimeTracker[string](time.Second)

	// t=0: insert foo, bar, buz; Get bar (no-op for lifetime); Evict foo.
	base.Op("foo", policy.Set)
	base.Op("bar", policy.Set)
	base.Op("buz", policy.Set)
	base.Op("bar", policy.Get)
	base.Op("foo", policy.Evict)

	// t=1.5s: bar (inserted before buz) should come first.
	fc.MoveBy(1500 * time.Millisecond)

	go base.cleanup()

	keys := collectN(t, base.ch, 2)
	assert.Equal(t, []string{"bar", "buz"}, keys)

	assert.Nil(t, base.Close())
}

func TestIdlePolicyHarness(t *testing.T) {
	policytest.RunTests(t, func(_ testing.TB) policy.EvictionPolicy[string] {
		return NewIdlePolicy[string](time.Hour)
	})
}

func TestLifetimePolicyHarness(t *testing.T) {
	policytest.RunTests(t, func(_ testing.TB) policy.EvictionPolicy[string] {
		return NewLifetimePolicy[string](time.Hour)
	})
}

func BenchmarkIdlePolicy(b *testing.B) {
	policytest.RunBenchmarks(b, func(_ testing.TB) policy.EvictionPolicy[string] {
		return NewIdlePolicy[string](time.Hour)
	})
}

func BenchmarkLifetimePolicy(b *testing.B) {
	policytest.RunBenchmarks(b, func(_ testing.TB) policy.EvictionPolicy[string] {
		return NewLifetimePolicy[string](time.Hour)
	})
}
