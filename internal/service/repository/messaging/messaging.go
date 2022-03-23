package messaging

import "github.com/inspii/comet/internal/service/model"

type SubscribeHandler func(topic string, message model.Message)

type Queue interface {
	// Subscribe 订阅消息，同一主题，各个订阅者都会收到消息
	Subscribe(topicPattern string, handler SubscribeHandler)
	// QueueSubscribe 订阅消息，同一主题，只有一个订阅者会收到消息
	QueueSubscribe(topicPattern string, handler SubscribeHandler)
	// Unsubscribe 取消订阅
	Unsubscribe(topicPattern string, handler SubscribeHandler)
	// Publish 发布消息
	Publish(topic string, msg model.Message) error
}
