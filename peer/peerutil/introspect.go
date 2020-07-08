package peerutil

import (
	"github.com/upfluence/pkg/log"
	"github.com/upfluence/pkg/peer"
	"github.com/upfluence/pkg/peer/version"
)

func Introspect(p *peer.Peer, ifs map[string]version.Version) {
	log.Noticef("Service %s %s", p.InstanceName, p.Version.String())

	if len(p.Interfaces) > 0 {
		log.Noticef("Interface versions:")

		for _, iface := range ifs {
			log.Noticef("* %s %s", iface.Name(), face.String())
		}
	}
}
