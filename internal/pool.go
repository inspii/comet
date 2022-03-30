package internal

import "sync"

type ListPeerOption struct {
	Limit int
}

type cometPool struct {
	peersMu    sync.RWMutex
	peers      map[string]Peer
	servicesMu sync.RWMutex
	services   map[string]Service
}

func (p *cometPool) GetPeer(id string) (Peer, bool) {
	p.peersMu.RLock()
	defer p.peersMu.RUnlock()

	peer, ok := p.peers[id]
	return peer, ok
}

func (p *cometPool) ListPeer(option ListPeerOption) []Peer {
	p.peersMu.RLock()
	defer p.peersMu.RUnlock()

	peers := make([]Peer, 0, len(p.peers))
	for _, peer := range p.peers {
		peers = append(peers, peer)
	}
	return peers
}

func (p *cometPool) AddPeer(peer Peer) error {
	p.peersMu.Lock()
	defer p.peersMu.Unlock()

	p.peers[peer.Info().ID] = peer
	return nil
}

func (p *cometPool) RemovePeer(id string) {
	p.peersMu.Lock()
	defer p.peersMu.Unlock()

	delete(p.peers, id)
}

func (p *cometPool) CountPeer() int {
	return len(p.peers)
}

func (p *cometPool) GetService(name string) (Service, bool) {
	p.servicesMu.RLock()
	defer p.servicesMu.RUnlock()

	service, ok := p.services[name]
	return service, ok
}

func (p *cometPool) ListService(option ListPeerOption) []Service {
	p.servicesMu.RLock()
	defer p.servicesMu.RUnlock()

	services := make([]Service, 0, len(p.services))
	for _, service := range p.services {
		services = append(services, service)
	}
	return services
}

func (p *cometPool) AddService(service Service) error {
	p.servicesMu.Lock()
	defer p.servicesMu.Unlock()

	p.services[service.Info().Name] = service
	return nil
}

func (p *cometPool) RemoveService(service Service) {
	p.servicesMu.Lock()
	defer p.servicesMu.Unlock()

	delete(p.services, service.Info().Name)
}

func (p *cometPool) CountService() int {
	return len(p.services)
}

type servicePool struct {
	sessionsMu sync.RWMutex
	sessions   map[string]ServiceWorker
}

func (p *servicePool) GetSession(id string) (ServiceWorker, bool) {
	p.sessionsMu.RLock()
	defer p.sessionsMu.RUnlock()

	peer, ok := p.sessions[id]
	return peer, ok
}

func (p *servicePool) ListSession(option ListPeerOption) []ServiceWorker {
	p.sessionsMu.RLock()
	defer p.sessionsMu.RUnlock()

	sessions := make([]ServiceWorker, 0, len(p.sessions))
	for _, session := range p.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (p *servicePool) AddSession(session ServiceWorker) error {
	p.sessionsMu.Lock()
	defer p.sessionsMu.Unlock()

	p.sessions[session.Info().ID] = session
	return nil
}

func (p *servicePool) RemoveSession(id string) {
	p.sessionsMu.Lock()
	defer p.sessionsMu.Unlock()

	delete(p.sessions, id)
}

func (p *servicePool) CountSession() int {
	return len(p.sessions)
}
