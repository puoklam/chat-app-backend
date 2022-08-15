package mq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/nsqio/go-nsq"
	"github.com/puoklam/chat-app-backend/api"
	"github.com/puoklam/chat-app-backend/env"
	"github.com/puoklam/chat-app-backend/ws"
)

const (
	SignalClearConsumers string = "SIGNAL_CLEAR_CONSUMERS"
	SignalAddConsumers   string = "SIGNAL_ADD_CONSUMERS"
)

type ExchangeMessage struct {
	Type          string          `json:"type"`
	UserID        uint            `json:"user_id"`
	TargetID      uint            `json:"target_id"`
	TargetType    string          `json:"target_type"`
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
			// User exit group
			remaining := clearConsumers(msg.UserID, msg.Topic)
			if remaining == 0 {
				// TODO: what to do if publish error
				Publish(env.EXCHANGE_NSQD_TCP_ADDR, msg.PostbackTopic, msg.PostbackMsg)
			}
		case SignalAddConsumers:
			// User create / join group
			clients := ws.GetHub().Clients(msg.UserID)
			for _, c := range clients {
				if c == nil {
					continue
				}
				topic := msg.Topic
				ch := c.Session().Ch
				csr, err := NewConsumer(topic, ch)
				if err != nil {
					continue
				}
				// declare new vars to store data to prevent handler dst, dsttype keep using reference
				// ?? maybe no need
				targetID, targetType := msg.TargetID, msg.TargetType
				csr.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
					var data BroadCastMessage
					if err := json.Unmarshal(message.Body, &data); err != nil {
						return err
					}
					m := api.OutMessage{
						ID:           string(message.ID[:]),
						FromID:       data.From.ID,
						FromName:     data.From.Displayname,
						FromImageURL: data.From.ImageURL,
						Dst:          targetID,
						DstType:      targetType,
						Content:      string(data.Body),
						Timestamp:    message.Timestamp,
					}
					b, err := json.Marshal(m)
					if err != nil {
						return err
					}
					msg := api.NewMessage(message, b)
					c.Send() <- msg
					return nil
				}))
				if csr.ConnectToNSQLookupd(env.NSQLOOKUPD_ADDR) != nil {
					csr.Stop()
					continue
				}
				if err := c.AddConsumer(context.Background(), topic, csr); err != nil {
					log.Println(err)
					csr.Stop()
				}
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
