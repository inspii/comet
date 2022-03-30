package internal

import (
	"errors"
	"io"
)

var (
	errPeerAlreadySetReceiveChannel = errors.New("peer already set receive channel")
	errPeerClosed                   = errors.New("peer closed")
)

type PeerInfo struct {
	ID              string          `json:"id"`
	Protocol        string          `json:"protocol"`        // 连接协议，目前只支持JSON
	ClientID        string          `json:"client_id"`       // 客户端ID
	IP              string          `json:"ip"`              // 客户端IP
	Service         string          `json:"service"`         // 业务系统，TODO 多业务系统支持
	ServiceToken    string          `json:"service_token"`   // 业务系统认证信息
	ServiceIdentity ServiceIdentity `json:"client_identity"` // 业务系统客户端信息
}

type Peer interface {
	Info() PeerInfo
	Receive(out chan<- *Message) error
	Send(msg *Message) error
	Subscribe()
	UnSubscribe()
}

type peerImpl struct {
	conn io.ReadWriter
	info PeerInfo
	in   chan *Message
	out  chan<- *Message
}

func NewPeer(conn io.ReadWriter, info PeerInfo) Peer {
	return &peerImpl{
		conn: conn,
		info: info,
	}
}

func (p *peerImpl) Info() PeerInfo {
	return p.info
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
