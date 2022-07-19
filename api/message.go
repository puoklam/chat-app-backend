package api

import (
	"github.com/nsqio/go-nsq"
	"github.com/puoklam/chat-app-backend/db/model"
)

type Message struct {
	*nsq.Message
	// logger  *log.Logger
	Content []byte
}

func (m *Message) Body() []byte {
	return m.Content
}

func (m *Message) OnError(err error) {
	// m.logger.Println(err)
	m.RequeueWithoutBackoff(0)
}

func (m *Message) OnSuccess() {
	m.Finish()
}

func NewMessage(m *nsq.Message, c []byte) *Message {
	return &Message{m, c}
}

type OutUser struct {
	model.Base
	Username    string `json:"username"`
	Displayname string `json:"displayname"`
}

type OutMessage struct {
	From      *OutUser `json:"from"`
	Dst       uint     `json:"dst"`
	DstType   string   `json:"dst_type"`
	Content   string   `json:"content"`
	Timestamp int64    `json:"timestamp"`
}
