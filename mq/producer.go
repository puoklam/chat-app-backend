package mq

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	// "sync"

	"github.com/nsqio/go-nsq"
	"github.com/puoklam/chat-app-backend/db/model"
	"github.com/puoklam/chat-app-backend/env"
)

// host -> producer
var producers struct {
	sync.Mutex
	m map[string]*nsq.Producer
}

// var pOnce sync.Once

type User struct {
	model.Base
	Username    string `json:"username"`
	Displayname string `json:"displayname"`
	ImageURL    string `json:"image_url"`
}

type Message interface {
	*BroadCastMessage | *ExchangeMessage | *PrivateMessage
}

type BroadCastMessage struct {
	From User
	Body []byte
}

func init() {
	producers.m = make(map[string]*nsq.Producer)
	cfg := nsq.NewConfig()
	addr := env.NSQD_TCP_ADDR
	p, err := nsq.NewProducer(addr, cfg)
	if err != nil {
		os.Exit(1)
	}
	producers.m[addr] = p
}

// func GetProducer() *nsq.Producer {
// 	pOnce.Do(func() {
// 		cfg := nsq.NewConfig()
// 		addr := os.Getenv("NSQD_TCP_ADDR")
// 		p, err := nsq.NewProducer(addr, cfg)
// 		if err != nil {
// 			os.Exit(1)
// 		}
// 		producer = p
// 	})
// 	return producer
// }

func insertHost(host string) error {
	producers.Lock()
	defer producers.Unlock()
	if _, ok := producers.m[host]; !ok {
		cfg := nsq.NewConfig()
		p, err := nsq.NewProducer(host, cfg)
		if err != nil {
			return err
		}
		producers.m[host] = p
	}
	return nil
}

func Publish[T Message](host string, topic string, message T) error {
	if !nsq.IsValidTopicName(topic) {
		return fmt.Errorf("invalid topic name: %s", topic)
	}
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}
	insertHost(host)
	return producers.m[host].Publish(topic, body)
}

func stopProducer(host string) {
	// producers.Lock()
	// defer producers.Unlock()
	if p, ok := producers.m[host]; ok {
		p.Stop()
		delete(producers.m, host)
	}
}

func StopProducers() {
	producers.Lock()
	defer producers.Unlock()
	for h := range producers.m {
		stopProducer(h)
	}
}
