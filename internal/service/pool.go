package service

type Pool interface {
	AddPeer(peer Peer) error
	RemovePeer(peer Peer) error
	RangePeers(f func(peer Peer) bool) error
}
