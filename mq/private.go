package mq

import (
	"encoding/json"
	"log"

	"github.com/nsqio/go-nsq"
	"github.com/puoklam/chat-app-backend/env"
)

const (
	SignalClearConsumersComplete string = "ccc"
)

type PrivateMessage struct {
	Type   string `json:"type"`
	UserID uint   `json:"user_id"`
	Topic  string `json:"topic"`
}

var pc *nsq.Consumer

func init() {
	consumer, err := nsq.NewConsumer(env.SERVER_ID, env.SERVER_ID, nsq.NewConfig())
	if err != nil {
		log.Fatalln(err)
	}
	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		var msg PrivateMessage
		if err := json.Unmarshal(message.Body, &msg); err != nil {
			// TODO: handle error
		}
		switch msg.Type {
		case SignalClearConsumersComplete:
			notifyClearComplete(msg.UserID, msg.Topic)
		}
		return nil
	}))
	if err := consumer.ConnectToNSQLookupd(env.NSQLOOKUPD_ADDR); err != nil {
		consumer.Stop()
		log.Fatalln(err)
	}
	pc = consumer
}

func notifyClearComplete(uid uint, topic string) {
	if topics := CleanUpChans[uid]; topics != nil {
		if ch := topics[topic]; ch != nil {
			ch <- true
		}
	}
}
