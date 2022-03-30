package internal

import (
	"fmt"
	"io"
)

type IndexEntry map[string]string

type ServiceIdentity struct {
	Identity    string      `json:"identity"`     // 业务系统用户唯一标识
	IndexedInfo IndexEntry  `json:"indexed_info"` // 业务系统用户信息（可搜索）
	ExtraInfo   interface{} `json:"extra_info"`   // 业务系统用户额外信息（不可搜索）
}

type ServiceWorkerInfo struct {
	ID string
}

type ServiceWorker interface {
	Info() ServiceWorkerInfo
	Receive(out chan<- *Message) error
	Send(msg *Message) error
}

func NewServiceWorker(conn io.ReadWriter) ServiceWorker {
	return nil
}

type ServiceInfo struct {
	Name string
}

func (s ServiceInfo) Topics() (publishTopic, subscribeTopic string) {
	pubTopic := fmt.Sprintf("$.service.%s.pub", s.Name)
	subTopic := fmt.Sprintf("$.service.%s.sub", s.Name)
	return pubTopic, subTopic
}

type Service interface {
	Info() ServiceInfo
	Auth(token string) (ServiceIdentity, error)
	GetPeerTopics(peer Peer) (publishTopic, subscribeTopic string)

	AddWorker(worker ServiceWorker) error
	RemoveWorker(worker ServiceWorker)
}

func NewService() Service {
	return nil
}

type serviceImpl struct {
	messaging   Messaging
	info        ServiceInfo
	servicePool servicePool
}

func (s *serviceImpl) Info() ServiceInfo {
	//TODO implement me
	panic("implement me")
}

func (s *serviceImpl) Auth(token string) (ServiceIdentity, error) {
	//TODO implement me
	panic("implement me")
}

func (s *serviceImpl) GetPeerTopics(peer Peer) (publishTopic, subscribeTopic string) {
	//TODO implement me
	panic("implement me")
}

func (s *serviceImpl) AddWorker(worker ServiceWorker) error {
	pubTopic, subTopic := s.Info().Topics()
	s.messaging.QueueSubscribe(subTopic, "default", func(topic string, message Message) {
		worker.Send(&message)
	})
	go func() {
		buf := make(chan *Message)
		if err := worker.Receive(buf); err != nil {
			return
		}

		for {
			select {
			case msg := <-buf:
				if err := s.messaging.Publish(pubTopic, *msg); err != nil {
					return
				}
			}
		}
	}()

	return s.servicePool.AddSession(worker)
}

func (s *serviceImpl) RemoveWorker(worker ServiceWorker) {
	s.servicePool.RemoveSession(worker.Info().ID)
}
