package peer

import "github.com/upfluence/pkg/metadata"

type Peer interface {
	Addr() string
	Metadata() metadata.Metadata
}
