package internal

import (
	cryptoRand "crypto/rand"
	"encoding/hex"
	"io"
	"math/rand"
)

type Message struct {
	ID      string
	Service string
	Topic   string
	Payload []byte
	Time    int
}

const (
	TopicAuth = "$.auth"
	TopicJoin = "$.join"
)

type MessageAuth struct {
	IP    string
	Token string
}

type MessageJoin struct {
	IP string
}

type SubscribeHandler func(topic string, message Message)

type Subscriber interface {
	// Unsubscribe 取消订阅
	Unsubscribe()
}

type Messaging interface {
	// Subscribe 订阅消息，同一主题，各个订阅者都会收到消息
	Subscribe(topicPattern string, handler SubscribeHandler) Subscriber
	// QueueSubscribe 订阅消息，同一主题，只有一个订阅者会收到消息
	QueueSubscribe(topicPattern string, queue string, handler SubscribeHandler) Subscriber
	// Publish 发布消息
	Publish(topic string, msg Message) error
}

type subscription struct {
	topicPattern string
	queue        string
	sid          string
	handler      SubscribeHandler
}

func genId() string {
	u := make([]byte, 16)
	io.ReadFull(cryptoRand.Reader, u)
	return hex.EncodeToString(u)
}

type standAloneSubscriber struct {
	sublist *Sublist
	sub     *subscription
}

func (s *standAloneSubscriber) Unsubscribe() {
	s.sublist.Remove(s.sub.topicPattern, s.sub)
}

type standAloneImpl struct {
	sublist *Sublist
}

func NewStandAloneMessaging() Messaging {
	return &standAloneImpl{sublist: NewSublist()}
}

func (m *standAloneImpl) Subscribe(topicPattern string, handler SubscribeHandler) Subscriber {
	return m.QueueSubscribe(topicPattern, "", handler)
}

func (m *standAloneImpl) QueueSubscribe(topicPattern string, queue string, handler SubscribeHandler) Subscriber {
	s := &subscription{
		topicPattern: topicPattern,
		sid:          genId(),
		queue:        queue,
		handler:      handler,
	}
	m.sublist.Insert(topicPattern, s)

	return &standAloneSubscriber{
		sublist: m.sublist,
		sub:     s,
	}
}

func (m *standAloneImpl) Publish(topic string, msg Message) error {
	sublist := m.sublist.Match(topic)

	var subs []*subscription
	queueSubs := make(map[string][]*subscription)
	for _, item := range sublist {
		s := item.(*subscription)
		if s.queue == "" {
			subs = append(subs, s)
		} else {
			queueSubs[s.queue] = append(queueSubs[s.queue], s)
		}
	}

	for _, l := range subs {
		go l.handler(topic, msg)
	}
	for _, l := range queueSubs {
		i := rand.Intn(len(l))
		go l[i].handler(topic, msg)
	}

	return nil
}
