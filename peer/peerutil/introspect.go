package peerutil

import (
	"github.com/upfluence/pkg/v2/log"
	"github.com/upfluence/pkg/v2/peer"
	"github.com/upfluence/pkg/v2/peer/version"
)

func Introspect(p *peer.Peer, ifaces map[string]version.Version) {
	log.Noticef("Service %s %s", p.InstanceName, p.Version.String())

	if len(ifaces) > 0 {
		log.Noticef("Interface versions:")

		for n, iface := range ifaces {
			log.Noticef("* %s %s", n, iface.String())
		}
	}
}
