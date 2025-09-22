package peer

type Peer interface {
	comparable
	Addr() string
}
