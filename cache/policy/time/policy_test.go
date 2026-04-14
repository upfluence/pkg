package time

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/pkg/v2/cache/policy"
)

// fakeClock lets tests advance time deterministically without sleeping.
type fakeClock struct {
	ns atomic.Int64
}

func (fc *fakeClock) now() int64 { return fc.ns.Load() }

func (fc *fakeClock) advance(d time.Duration) { fc.ns.Add(int64(d)) }

// withFakeClock injects a fake clock into a Policy.
func withFakeClock[K comparable](p *Policy[K]) (*Policy[K], *fakeClock) {
	fc := &fakeClock{}
	p.now = fc.now
	return p, fc
}

// collectN reads n keys from ch, returning them in order.
// It runs concurrently so that the sender (cleanup) is not blocked.
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
	base, fc := withFakeClock(NewIdlePolicy[string](time.Second))

	// t=0: insert foo, bar, buz; Get foo; Evict bar.
	base.Op("foo", policy.Set)
	base.Op("bar", policy.Set)
	base.Op("buz", policy.Set)
	base.Op("foo", policy.Get)
	base.Op("bar", policy.Evict)

	// Advance to t=1.5s — both buz and foo are idle for >1s.
	// buz (never accessed after insert) is at front; foo (Get at t=0) is at back.
	fc.advance(1500 * time.Millisecond)

	// Run cleanup in a goroutine; collectN drains the channel concurrently.
	go base.cleanup(fc.now())

	keys := collectN(t, base.C(), 2)
	assert.Equal(t, []string{"buz", "foo"}, keys)

	assert.Nil(t, base.Close())
}

func TestIdlePolicyTimestampRefresh(t *testing.T) {
	// Verifies bug fix #2: Get on idle policy must refresh the timestamp so
	// the entry is NOT evicted at t=ttl when it was recently accessed.
	base, fc := withFakeClock(NewIdlePolicy[string](time.Second))

	base.Op("foo", policy.Set) // inserted at t=0

	// t=0.8s: Get foo — refreshes its idle timer to t=0.8s.
	fc.advance(800 * time.Millisecond)
	base.Op("foo", policy.Get)

	// t=1.2s: foo idle for only 0.4s — must NOT be evicted.
	fc.advance(400 * time.Millisecond)
	base.cleanup(fc.now())

	select {
	case k := <-base.C():
		t.Errorf("unexpected eviction of %q at t=1.2s (idle only 0.4s)", k)
	default:
	}

	// t=1.9s: foo idle for 1.1s > 1s — must be evicted.
	fc.advance(700 * time.Millisecond)
	go base.cleanup(fc.now())

	keys := collectN(t, base.C(), 1)
	assert.Equal(t, []string{"foo"}, keys)

	assert.Nil(t, base.Close())
}

func TestLifetimePolicy(t *testing.T) {
	base, fc := withFakeClock(NewLifetimePolicy[string](time.Second))

	// t=0: insert foo, bar, buz; Get bar (no-op for lifetime); Evict foo.
	base.Op("foo", policy.Set)
	base.Op("bar", policy.Set)
	base.Op("buz", policy.Set)
	base.Op("bar", policy.Get)
	base.Op("foo", policy.Evict)

	// t=1.5s: bar (inserted before buz) should come first.
	fc.advance(1500 * time.Millisecond)

	go base.cleanup(fc.now())

	keys := collectN(t, base.C(), 2)
	assert.Equal(t, []string{"bar", "buz"}, keys)

	assert.Nil(t, base.Close())
}
