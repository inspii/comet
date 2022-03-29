package internal

import (
	"fmt"
	"testing"
	"time"
)

func TestMessaging(t *testing.T) {
	m := NewStandAloneMessaging()
	m.Subscribe("a.*", func(topic string, message Message) {
		fmt.Println(topic, message.Payload)
	})
	m.Subscribe("a.*", func(topic string, message Message) {
		fmt.Println(topic, message.Payload)
	})
	m.Subscribe("a.*.*", func(topic string, message Message) {
		fmt.Println(topic, message.Payload)
	})

	m.Publish("a.b.c", Message{Payload: []byte("ac")})
	m.Publish("a.b", Message{Payload: []byte("ab")})
	time.Sleep(2 * time.Second)
}
