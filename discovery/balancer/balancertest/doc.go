// Package balancertest provides testing utilities for balancer.Policy implementations.
//
// Usage:
//
//	func TestMyPolicy(t *testing.T) {
//		balancertest.PolicyTest(t, func() balancer.Policy[static.Peer, static.Peer] {
//			return NewPolicy[static.Peer]()
//		})
//	}
//
// PolicyTest runs a comprehensive test suite that covers:
//   - Empty policy (no peers) - tests NoWait and context cancellation behavior
//   - Single peer - verifies Get() returns the same peer consistently
//   - Adding and removing peers - tests dynamic peer updates
//   - Removing all peers - verifies transition back to empty state
//   - Adding peers to empty policy - tests that waiting Get() calls unblock
package balancertest
