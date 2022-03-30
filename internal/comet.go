package internal

import (
	"errors"
	"io"
)

var (
	errServiceNotAvailable = errors.New("service not available")
)

type Comet struct {
	messaging Messaging
	pool      cometPool
}

func (c *Comet) NewPeer(conn io.ReadWriter, info PeerInfo) (Peer, error) {
	service, ok := c.pool.GetService(info.Service)
	if !ok {
		return nil, errServiceNotAvailable
	}
	identity, err := service.Auth(info.ServiceToken)
	if err != nil {
		return nil, err
	}

	info.ServiceIdentity = identity
	peer := NewPeer(conn, info)
	return peer, nil
}

func (c *Comet) AddPeer(peer Peer) error {
	service, ok := c.pool.GetService(peer.Info().Service)
	if !ok {
		return errServiceNotAvailable
	}
	pubTopic, subTopic := service.GetPeerTopics(peer)
	// TODO Unsubscribe
	c.messaging.Subscribe(subTopic, func(topic string, message Message) {
		peer.Send(&message)
	})
	go func() {
		buf := make(chan *Message)
		if err := peer.Receive(buf); err != nil {
			return
		}

		for {
			select {
			case msg := <-buf:
				if err := c.messaging.Publish(pubTopic, *msg); err != nil {
					return
				}
			}
		}
	}()

	return c.pool.AddPeer(peer)
}

func (c *Comet) RemovePeer(peer Peer) {
	c.pool.RemovePeer(peer.Info().ID)
}

func (c *Comet) ListPeer(option ListPeerOption) []Peer {
	return c.pool.ListPeer(option)
}

func (c *Comet) CountPeer() int {
	return c.pool.CountPeer()
}

func (c *Comet) NewService(service string) (Service, error) {
	sess := NewService()
	return sess, nil
}

func (c *Comet) NewServiceWorker(conn io.ReadWriter) (ServiceWorker, error) {
	sess := NewServiceWorker(conn)
	return sess, nil
}

func (c *Comet) GetService(name string) (Service, bool) {
	return c.pool.GetService(name)
}

func (c *Comet) ListService(option ListPeerOption) []Service {
	return c.pool.ListService(option)
}

func (c *Comet) CountService() int {
	return c.pool.CountService()
}

func (c *Comet) RegisterService(service Service) error {
	panic("not implemented")
}

func (c *Comet) UnregisterService(service Service) {
	panic("not implemented")
}
