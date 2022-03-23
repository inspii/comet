package service

type ListPeerOption struct {
	Limit int
}

type Comet interface {
	AddPeer(peer Peer) error
	CountPeer() (total int, err error)
	ListPeers(option ListPeerOption) (peer []Peer, err error)
}