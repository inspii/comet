package model

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
