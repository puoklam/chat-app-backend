package mq

import (
	"encoding/json"
	"log"

	"github.com/nsqio/go-nsq"
	"github.com/puoklam/chat-app-backend/env"
	"github.com/puoklam/chat-app-backend/ws"
)

const (
	SignalClearConsumers string = "cc"
)

type ExchangeMessage struct {
	Type          string          `json:"type"`
	UserID        uint            `json:"user_id"`
	Topic         string          `json:"topic"`
	PostbackTopic string          `json:"postback_topic"`
	PostbackCh    string          `json:"postback_ch"`
	PostbackMsg   *PrivateMessage `json:"postback_message"`
}

var ec *nsq.Consumer

func init() {
	consumer, err := nsq.NewConsumer("info", env.SERVER_ID, nsq.NewConfig())
	if err != nil {
		log.Fatalln(err)
	}
	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		var msg ExchangeMessage
		if err := json.Unmarshal(message.Body, &msg); err != nil {
			// TODO: handle error
		}
		switch msg.Type {
		case SignalClearConsumers:
			remaining := clearConsumers(msg.UserID, msg.Topic)
			if remaining == 0 {
				// TODO: what to do if publish error
				Publish(env.EXCHANGE_NSQD_TCP_ADDR, msg.PostbackTopic, msg.PostbackMsg)
			}
		}
		return nil
	}))
	if err := consumer.ConnectToNSQLookupd(env.NSQLOOKUPD_ADDR); err != nil {
		consumer.Stop()
		log.Fatalln(err)
	}
	ec = consumer
}

func clearConsumers(uid uint, topic string) int {
	count := -1
	clients := ws.GetHub().Clients(uid)
	for _, c := range clients {
		if c == nil {
			continue
		}
		count = c.StopConsumers(topic)
	}
	return count
}
