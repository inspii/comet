package service

import "github.com/inspii/comet/internal/service/model"

type Peer interface {
	Receive(out chan<- *model.Message) error
	Send(msg *model.Message) error
	Close()
	IsClosed() bool
	Subscribe()
	UnSubscribe()
}
