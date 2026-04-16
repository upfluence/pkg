// Package resolvertest provides reusable test helpers for resolver implementations.
//
// The ResolverTest function runs a comprehensive test suite that validates common
// resolver behaviors including:
//   - Empty resolver (no initial peers)
//   - Initial peers resolution
//   - NoWait behavior with no updates
//   - Context cancellation
//   - Watcher closure
//
// Example usage:
//
//	func TestMyResolver(t *testing.T) {
//		resolvertest.ResolverTest(t, func() resolver.Resolver[static.Peer] {
//			return myresolver.New()
//		})
//	}
package resolvertest
