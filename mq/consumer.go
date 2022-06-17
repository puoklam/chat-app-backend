package mq

import (
	"github.com/nsqio/go-nsq"
)

type Handler struct{}

func (h *Handler) HandleMessage(m *nsq.Message) error {
	// fmt.Println(string(m.Body))
	return nil
}

func NewConsumer(topic, ch string) (*nsq.Consumer, error) {
	cfg := nsq.NewConfig()
	return nsq.NewConsumer(topic, ch, cfg)
}
