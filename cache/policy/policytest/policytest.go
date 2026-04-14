// Package policytest provides a reusable test harness and benchmarks for
// implementations of [policy.EvictionPolicy].
//
// Usage in a package test:
//
//	func TestMyPolicy(t *testing.T) {
//	    policytest.RunTests(t, func(t testing.TB) policy.EvictionPolicy[string] {
//	        return mypackage.NewMyPolicy[string](...)
//	    })
//	}
//
//	func BenchmarkMyPolicy(b *testing.B) {
//	    policytest.RunBenchmarks(b, func(b *testing.B) policy.EvictionPolicy[string] {
//	        return mypackage.NewMyPolicy[string](...)
//	    })
//	}
package policytest

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/upfluence/pkg/v2/cache/policy"
)

// Factory creates a fresh EvictionPolicy[string] for each sub-test or
// sub-benchmark. The returned policy must not have been used yet.
type Factory func(testing.TB) policy.EvictionPolicy[string]

// RunTests runs the full correctness suite against the policy produced by f.
func RunTests(t *testing.T, f Factory) {
	t.Helper()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.fn(t, f(t))
		})
	}
}

// RunBenchmarks runs all benchmark workloads against the policy produced by f.
// f is called once per benchmark invocation so each run gets a fresh policy.
func RunBenchmarks(b *testing.B, f Factory) {
	b.Helper()

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			// The bench function is responsible for calling f to create a
			// policy. It receives f so it can also allocate additional
			// policies (e.g. close_under_load).
			bc.fn(b, f)
		})
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// drain reads from ch until it is closed or the deadline expires, returning
// all received keys.
func drain(t testing.TB, ch <-chan string, deadline time.Duration) []string {
	t.Helper()

	timer := time.NewTimer(deadline)
	defer timer.Stop()

	var out []string

	for {
		select {
		case k, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, k)
		case <-timer.C:
			return out
		}
	}
}

// receiveOne waits up to deadline for a single value from ch.
func receiveOne(t testing.TB, ch <-chan string, deadline time.Duration) (string, bool) {
	t.Helper()

	timer := time.NewTimer(deadline)
	defer timer.Stop()

	select {
	case k, ok := <-ch:
		return k, ok
	case <-timer.C:
		return "", false
	}
}

// assertChannelClosed asserts that ch is closed within the deadline.
func assertChannelClosed(t testing.TB, ch <-chan string, deadline time.Duration) {
	t.Helper()

	timer := time.NewTimer(deadline)
	defer timer.Stop()

	select {
	case _, ok := <-ch:
		assert.False(t, ok, "expected channel to be closed, got a value instead")
	case <-timer.C:
		t.Error("timed out waiting for channel to be closed")
	}
}

// assertChannelEmpty asserts that ch has no pending value right now.
func assertChannelEmpty(t testing.TB, ch <-chan string) {
	t.Helper()

	select {
	case k, ok := <-ch:
		if ok {
			t.Errorf("expected empty channel, got key %q", k)
		} else {
			t.Error("expected empty channel, channel was closed")
		}
	default:
	}
}

const shortWait = 50 * time.Millisecond

// ---------------------------------------------------------------------------
// correctness test cases
// ---------------------------------------------------------------------------

type testCase struct {
	name string
	fn   func(*testing.T, policy.EvictionPolicy[string])
}

var testCases = []testCase{
	{
		name: "close_empty_closes_channel",
		fn: func(t *testing.T, p policy.EvictionPolicy[string]) {
			// An unused policy: Close() must close the channel returned by C().
			ch := p.C()
			require.NoError(t, p.Close())
			assertChannelClosed(t, ch, shortWait)
		},
	},
	{
		name: "op_after_close_returns_err_closed",
		fn: func(t *testing.T, p policy.EvictionPolicy[string]) {
			require.NoError(t, p.Close())

			assert.Equal(t, policy.ErrClosed, p.Op("foo", policy.Set))
			assert.Equal(t, policy.ErrClosed, p.Op("foo", policy.Get))
			assert.Equal(t, policy.ErrClosed, p.Op("foo", policy.Evict))
		},
	},
	{
		name: "double_close_is_safe",
		fn: func(t *testing.T, p policy.EvictionPolicy[string]) {
			// Neither call must panic or deadlock.
			assert.NoError(t, p.Close())
			assert.NoError(t, p.Close())
		},
	},
	{
		name: "explicit_evict_suppresses_channel_send",
		fn: func(t *testing.T, p policy.EvictionPolicy[string]) {
			// Set a key then immediately evict it. The policy must remove the
			// key from its internal tracking; no eviction event should arrive
			// on C() for this key.
			require.NoError(t, p.Op("foo", policy.Set))
			require.NoError(t, p.Op("foo", policy.Evict))

			assertChannelEmpty(t, p.C())
			require.NoError(t, p.Close())
		},
	},
	{
		name: "set_same_key_twice_no_duplicate",
		fn: func(t *testing.T, p policy.EvictionPolicy[string]) {
			// Setting the same key twice must not create a duplicate entry.
			// No channel send should result from the second Set alone.
			require.NoError(t, p.Op("foo", policy.Set))
			require.NoError(t, p.Op("foo", policy.Set))

			assertChannelEmpty(t, p.C())
			require.NoError(t, p.Close())
		},
	},
	{
		name: "get_on_unknown_key_is_noop",
		fn: func(t *testing.T, p policy.EvictionPolicy[string]) {
			// A Get for a key never Set must not panic, must return nil, and
			// must not produce a channel event.
			require.NoError(t, p.Op("ghost", policy.Get))

			assertChannelEmpty(t, p.C())
			require.NoError(t, p.Close())
		},
	},
	{
		name: "evict_on_unknown_key_is_noop",
		fn: func(t *testing.T, p policy.EvictionPolicy[string]) {
			// An Evict for a key never Set must not panic, must return nil,
			// and must not produce a channel event.
			require.NoError(t, p.Op("ghost", policy.Evict))

			assertChannelEmpty(t, p.C())
			require.NoError(t, p.Close())
		},
	},
	{
		name: "channel_closed_after_close",
		fn: func(t *testing.T, p policy.EvictionPolicy[string]) {
			// Obtain the channel before closing so we test the same instance.
			ch := p.C()
			require.NoError(t, p.Close())
			assertChannelClosed(t, ch, shortWait)
		},
	},
	{
		name: "channel_closed_after_close_obtained_post_close",
		fn: func(t *testing.T, p policy.EvictionPolicy[string]) {
			// C() called after Close() must still return a closed channel.
			require.NoError(t, p.Close())
			ch := p.C()
			assertChannelClosed(t, ch, shortWait)
		},
	},
	{
		name: "concurrent_ops_and_close_no_race",
		fn: func(t *testing.T, p policy.EvictionPolicy[string]) {
			// Hammer Op from many goroutines while Close is called concurrently.
			// The race detector is the real assertion here; we just need no
			// panic and no deadlock.
			const goroutines = 8
			const opsEach = 200

			keys := [4]string{"a", "b", "c", "d"}

			var wg sync.WaitGroup

			// Drain the channel so senders are never blocked indefinitely.
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range p.C() {
				}
			}()

			// Writers.
			for i := range goroutines {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					for j := range opsEach {
						op := policy.OpType(j % 3)  // Set, Get, Evict
						p.Op(keys[i%len(keys)], op) //nolint:errcheck
					}
				}(i)
			}

			// Give writers a head-start then close.
			time.Sleep(2 * time.Millisecond)
			require.NoError(t, p.Close())

			wg.Wait()
		},
	},
	{
		name: "concurrent_close_calls_no_race",
		fn: func(t *testing.T, p policy.EvictionPolicy[string]) {
			// Multiple goroutines closing simultaneously must not panic or race.
			// Drain so senders don't block.
			go func() {
				for range p.C() {
				}
			}()

			var wg sync.WaitGroup

			for range 8 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					p.Close() //nolint:errcheck
				}()
			}

			wg.Wait()
		},
	},
	{
		name: "c_returns_same_channel",
		fn: func(t *testing.T, p policy.EvictionPolicy[string]) {
			// Repeated calls to C() must return the same channel instance so
			// consumers can safely cache the result.
			ch1 := p.C()
			ch2 := p.C()
			assert.Equal(t, ch1, ch2, "C() must return the same channel on repeated calls")
			require.NoError(t, p.Close())
		},
	},
}

// ---------------------------------------------------------------------------
// benchmark workloads
// ---------------------------------------------------------------------------

// benchFn is the signature for a benchmark case. It receives f so it can
// allocate fresh policies as needed (including per-iteration in
// close_under_load).
type benchCase struct {
	name string
	fn   func(*testing.B, Factory)
}

var benchCases = []benchCase{
	{
		// Pure Set pressure with a small working set so LRU-style policies
		// produce frequent evictions.
		name: "set_only",
		fn: func(b *testing.B, f Factory) {
			p := f(b)
			defer p.Close() //nolint:errcheck
			keys := makeBenchKeys(64)
			drainAsync(b, p)
			b.ResetTimer()

			for i := range b.N {
				p.Op(keys[i%len(keys)], policy.Set) //nolint:errcheck
			}
		},
	},
	{
		// Read-heavy: 90 % Get, 10 % Set. Models a warm cache.
		name: "get_heavy",
		fn: func(b *testing.B, f Factory) {
			p := f(b)
			defer p.Close() //nolint:errcheck
			keys := makeBenchKeys(64)

			// Drain must start before the pre-warm loop so that evictions
			// produced by inserting more keys than the policy's capacity do
			// not block on the channel.
			drainAsync(b, p)

			for _, k := range keys {
				p.Op(k, policy.Set) //nolint:errcheck
			}

			b.ResetTimer()

			for i := range b.N {
				if i%10 == 0 {
					p.Op(keys[i%len(keys)], policy.Set) //nolint:errcheck
				} else {
					p.Op(keys[i%len(keys)], policy.Get) //nolint:errcheck
				}
			}
		},
	},
	{
		// Write-heavy: 90 % Set, 10 % Get. High eviction rate for size-based
		// policies.
		name: "write_heavy",
		fn: func(b *testing.B, f Factory) {
			p := f(b)
			defer p.Close() //nolint:errcheck
			keys := makeBenchKeys(64)
			drainAsync(b, p)
			b.ResetTimer()

			for i := range b.N {
				if i%10 == 0 {
					p.Op(keys[i%len(keys)], policy.Get) //nolint:errcheck
				} else {
					p.Op(keys[i%len(keys)], policy.Set) //nolint:errcheck
				}
			}
		},
	},
	{
		// Mixed: 50 % Set, 50 % Get.
		name: "mixed",
		fn: func(b *testing.B, f Factory) {
			p := f(b)
			defer p.Close() //nolint:errcheck
			keys := makeBenchKeys(64)

			// Drain must start before the pre-warm loop (same reason as
			// get_heavy: pre-warming more keys than the policy capacity would
			// otherwise block on the eviction channel).
			drainAsync(b, p)

			for _, k := range keys {
				p.Op(k, policy.Set) //nolint:errcheck
			}

			b.ResetTimer()

			for i := range b.N {
				if i%2 == 0 {
					p.Op(keys[i%len(keys)], policy.Set) //nolint:errcheck
				} else {
					p.Op(keys[i%len(keys)], policy.Get) //nolint:errcheck
				}
			}
		},
	},
	{
		// Evict-heavy: alternating Set / Evict — stresses the remove path and
		// ensures the policy's internal bookkeeping stays consistent under
		// churn.
		name: "evict_heavy",
		fn: func(b *testing.B, f Factory) {
			p := f(b)
			defer p.Close() //nolint:errcheck
			keys := makeBenchKeys(64)
			drainAsync(b, p)
			b.ResetTimer()

			for i := range b.N {
				k := keys[i%len(keys)]
				if i%2 == 0 {
					p.Op(k, policy.Set) //nolint:errcheck
				} else {
					p.Op(k, policy.Evict) //nolint:errcheck
				}
			}
		},
	},
	{
		// Concurrent: GOMAXPROCS goroutines each running a mixed workload.
		// Use -cpu 1,2,4,8 to observe scalability.
		name: "concurrent",
		fn: func(b *testing.B, f Factory) {
			p := f(b)
			defer p.Close() //nolint:errcheck
			keys := makeBenchKeys(64)
			drainAsync(b, p)
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				var i int
				for pb.Next() {
					k := keys[i%len(keys)]
					switch i % 3 {
					case 0:
						p.Op(k, policy.Set) //nolint:errcheck
					case 1:
						p.Op(k, policy.Get) //nolint:errcheck
					case 2:
						p.Op(k, policy.Evict) //nolint:errcheck
					}
					i++
				}
			})
		},
	},
	{
		// Close under load: measures the cost of Close() while concurrent Op
		// calls are still in flight. Each b.N iteration allocates its own
		// fresh policy via the factory so that Close() is meaningful every
		// time.
		name: "close_under_load",
		fn: func(b *testing.B, f Factory) {
			keys := makeBenchKeys(64)
			b.ResetTimer()

			for range b.N {
				pp := f(b)

				go func() {
					for range pp.C() {
					}
				}()

				done := make(chan struct{})
				var wg sync.WaitGroup

				for range 4 {
					wg.Add(1)
					go func() {
						defer wg.Done()
						var i int
						for {
							select {
							case <-done:
								return
							default:
								pp.Op(keys[i%len(keys)], policy.Set) //nolint:errcheck
								i++
							}
						}
					}()
				}

				time.Sleep(100 * time.Microsecond)
				close(done)
				pp.Close() //nolint:errcheck
				wg.Wait()
			}
		},
	},
}

// makeBenchKeys returns n distinct string keys for benchmark use.
func makeBenchKeys(n int) []string {
	keys := make([]string, n)
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_-"
	for i := range keys {
		// Two-character key from the alphabet — sufficient for n <= 64*64.
		keys[i] = string([]byte{alphabet[i/len(alphabet)], alphabet[i%len(alphabet)]})
	}
	return keys
}

// drainAsync starts a goroutine that drains p.C() until it is closed, so that
// sending policies are never blocked on the channel during benchmarks.
func drainAsync(b *testing.B, p policy.EvictionPolicy[string]) {
	b.Helper()

	go func() {
		for range p.C() {
		}
	}()
}
