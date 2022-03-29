package internal

import (
	"errors"
	"io"
)

type Peer interface {
	Receive(out chan<- *Message) error
	Send(msg *Message) error
	Subscribe()
	UnSubscribe()
}

var (
	errPeerAlreadySetReceiveChannel = errors.New("peer already set receive channel")
	errPeerClosed                   = errors.New("peer closed")
)

type peerImpl struct {
	conn io.ReadWriter
	in   chan *Message
	out  chan<- *Message
}

func NewPeer(conn io.ReadWriter) Peer {
	return &peerImpl{
		conn: conn,
	}
}

func (p *peerImpl) Receive(out chan<- *Message) error {
	if p.out != nil {
		return errPeerAlreadySetReceiveChannel
	}
	p.out = out
	return nil
}

func (p *peerImpl) Send(msg *Message) error {
	p.in <- msg
	return nil
}

func (p *peerImpl) Subscribe() {
	//TODO implement me
	panic("implement me")
}

func (p *peerImpl) UnSubscribe() {
	//TODO implement me
	panic("implement me")
}
