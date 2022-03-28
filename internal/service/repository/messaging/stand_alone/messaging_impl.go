package stand_alone

import (
	cryptoRand "crypto/rand"
	"encoding/hex"
	"github.com/inspii/comet/internal/service/model"
	"github.com/inspii/comet/internal/service/repository/messaging"
	"io"
	"math/rand"
)

type subscription struct {
	topicPattern string
	queue        string
	sid          string
	handler      messaging.SubscribeHandler
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

func NewStandAloneMessaging() messaging.Messaging {
	return &standAloneImpl{sublist: New()}
}

func (m *standAloneImpl) Subscribe(topicPattern string, handler messaging.SubscribeHandler) messaging.Subscriber {
	return m.QueueSubscribe(topicPattern, "", handler)
}

func (m *standAloneImpl) QueueSubscribe(topicPattern string, queue string, handler messaging.SubscribeHandler) messaging.Subscriber {
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

func (m *standAloneImpl) Publish(topic string, msg model.Message) error {
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
