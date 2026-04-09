// Package simple provides a flexible balancer policy that delegates peer
// selection to a Picker implementation.
//
// The Picker interface allows for custom balancing strategies without
// implementing the full Policy interface:
//
//	type CustomPicker struct{}
//
//	func (p *CustomPicker) Pick(ctx context.Context, peers []peer.Peer) (peer.Peer, error) {
//	    // Custom selection logic
//	    return peers[0], nil
//	}
//
//	policy := simple.NewPolicy[*peer.Peer](new(CustomPicker))
package simple
