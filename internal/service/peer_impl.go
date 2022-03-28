package service

import (
	"errors"
	"github.com/inspii/comet/internal/service/model"
	"io"
)

var (
	errPeerAlreadySetReceiveChannel = errors.New("peer already set receive channel")
	errPeerClosed                   = errors.New("peer closed")
)

type peerImpl struct {
	rw  io.ReadWriter
	in  chan *model.Message
	out chan<- *model.Message
}

func NewPeer(rw io.ReadWriter) Peer {
	return &peerImpl{
		rw: rw,
	}
}

func (p *peerImpl) Receive(out chan<- *model.Message) error {
	if p.out != nil {
		return errPeerAlreadySetReceiveChannel
	}
	p.out = out
	return nil
}

func (p *peerImpl) Send(msg *model.Message) error {
	if p.IsClosed() {
		return errPeerClosed
	}

	p.in <- msg
	return nil
}

func (p *peerImpl) Close() {
	//TODO implement me
	panic("implement me")
}

func (p *peerImpl) IsClosed() bool {
	//TODO implement me
	panic("implement me")
}

func (p *peerImpl) Subscribe() {
	//TODO implement me
	panic("implement me")
}

func (p *peerImpl) UnSubscribe() {
	//TODO implement me
	panic("implement me")
}
