package mq

import (
	"encoding/json"
	"errors"
	"os"

	// "sync"

	"github.com/nsqio/go-nsq"
	"github.com/puoklam/chat-app-backend/db/model"
)

var producer *nsq.Producer

// var pOnce sync.Once

type User struct {
	model.Base
	Username    string `json:"username"`
	Displayname string `json:"displayname"`
}

type Message struct {
	From User
	Body []byte
}

func init() {
	cfg := nsq.NewConfig()
	addr := os.Getenv("NSQD_ADDR")
	p, err := nsq.NewProducer(addr, cfg)
	if err != nil {
		os.Exit(1)
	}
	producer = p
}

// func GetProducer() *nsq.Producer {
// 	pOnce.Do(func() {
// 		cfg := nsq.NewConfig()
// 		addr := os.Getenv("NSQD_ADDR")
// 		p, err := nsq.NewProducer(addr, cfg)
// 		if err != nil {
// 			os.Exit(1)
// 		}
// 		producer = p
// 	})
// 	return producer
// }

func Publish(topic string, message *Message) error {
	if !nsq.IsValidTopicName(topic) {
		return errors.New("invalid topic name")
	}
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return producer.Publish(topic, body)
}

func StopProducer() {
	producer.Stop()
}
