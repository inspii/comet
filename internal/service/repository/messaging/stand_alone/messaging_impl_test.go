package stand_alone

import (
	"fmt"
	"github.com/inspii/comet/internal/service/model"
	"testing"
	"time"
)

func TestMessaging(t *testing.T) {
	m := NewStandAloneMessaging()
	m.Subscribe("a.*", func(topic string, message model.Message) {
		fmt.Println(topic, message.Payload)
	})
	m.Subscribe("a.*", func(topic string, message model.Message) {
		fmt.Println(topic, message.Payload)
	})
	m.Subscribe("a.*.*", func(topic string, message model.Message) {
		fmt.Println(topic, message.Payload)
	})

	m.Publish("a.b.c", model.Message{Payload: "ac"})
	m.Publish("a.b", model.Message{Payload: "ab"})
	time.Sleep(2 * time.Second)
}
