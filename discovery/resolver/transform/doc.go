// Package transform provides a resolver wrapper that transforms peers from one type to another.
//
// The transform resolver wraps an existing resolver and applies a transformation function
// to convert source peers (S) into target peers (T). This is useful for:
//   - Adding metadata or wrapping peers with additional information
//   - Converting between different peer implementations
//   - Applying prefixes or transformations to peer addresses
//
// Example:
//
//	source := static.NewResolverFromStrings([]string{"host1:80", "host2:80"})
//
//	// Wrap peers with a prefix
//	tr := transform.WrapResolver(source, func(p static.Peer) wrappedPeer {
//		return wrappedPeer{addr: p.Addr(), prefix: "service-"}
//	})
//
//	// The transformed resolver will emit wrappedPeer instances
//	// with addresses like "service-host1:80", "service-host2:80"
package transform
