package connectionpicker

import (
	"fmt"

	"github.com/upfluence/pkg/cfg"
	"github.com/upfluence/pkg/discovery/balancer"
	"github.com/upfluence/pkg/discovery/balancer/random"
	"github.com/upfluence/pkg/discovery/resolver/static"
	"github.com/upfluence/pkg/peer"
)

var (
	defaultOptions = &options{
		Balancer:         defaultBalancer(),
		peer:             peer.FromEnv(),
		targetOpenedConn: 1,
		connectionNamer: func(p *peer.Peer, d int) string {
			return fmt.Sprintf("%s-%d", p.URL().String(), d)
		},
	}
)

func defaultBalancer() balancer.Balancer {
	var uris = cfg.FetchStrings(
		"RABBITMQ_URLS",
		[]string{cfg.FetchString("RABBITMQ_URL", "localhost:5672")},
	)

	return random.NewBalancer(static.NewResolverFromStrings(uris))
}

type Option func(*options)

func WithBalancer(b balancer.Balancer) Option {
	return func(o *options) { o.Balancer = b }
}

func WithPeer(p *peer.Peer) Option { return func(o *options) { o.peer = p } }

func WithConnectionNameTemplate(tmpl string) Option {
	return func(o *options) {
		o.connectionNamer = func(p *peer.Peer, v int) string { return fmt.Sprintf(tmpl, p, v) }
	}
}

func MaxOpenedConnections(v int) Option {
	return func(o *options) { o.targetOpenedConn = v }
}

type options struct {
	balancer.Balancer

	targetOpenedConn int

	peer            *peer.Peer
	connectionNamer func(*peer.Peer, int) string
}
