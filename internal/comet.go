package internal

import "sync"

type ListPeerOption struct {
	Limit int
}

type Comet struct {
	messaging Messaging
	peersMu   sync.RWMutex
	peers     map[Peer]struct{}
}

func (c *Comet) AddPeer(peer Peer) error {
	c.peersMu.Lock()
	defer c.peersMu.Unlock()

	c.peers[peer] = struct{}{}
	return nil
}

func (c *Comet) RemovePeer(peer Peer) {
	delete(c.peers, peer)
}

func (c *Comet) ListPeer(option ListPeerOption) []Peer {
	peers := make([]Peer, 0, len(c.peers))
	for peer := range c.peers {
		peers = append(peers, peer)
	}
	return peers
}

func (c *Comet) CountPeer() int {
	return len(c.peers)
}
