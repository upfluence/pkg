package peerutil

import (
	"github.com/upfluence/pkg/log"
	"github.com/upfluence/pkg/peer"
)

func Introspect(p *peer.Peer) {
	log.Noticef("Service %s %s", p.InstanceName, peer.SerializeVersion(p.Version))

	if len(p.Interfaces) > 0 {
		log.Noticef("Interface versions:")

		for _, iface := range p.Interfaces {
			log.Noticef("* %s %s", iface.Name(), peer.SerializeVersion(iface.Version()))
		}
	}
}
