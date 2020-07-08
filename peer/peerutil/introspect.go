package peerutil

import (
	"github.com/upfluence/pkg/log"
	"github.com/upfluence/pkg/peer"
	"github.com/upfluence/pkg/peer/version"
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
