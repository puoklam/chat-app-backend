package mq

import (
	"os"
	"sync"

	"github.com/nsqio/go-nsq"
)

var producer *nsq.Producer
var once sync.Once

type Message struct {
	Body      string
	Timestamp int64
}

func GetProducer() *nsq.Producer {
	once.Do(func() {
		cfg := nsq.NewConfig()
		p, err := nsq.NewProducer("localhost:4150", cfg)
		if err != nil {
			os.Exit(1)
		}
		producer = p
	})
	return producer
}
