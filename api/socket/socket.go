package socket

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/nsqio/go-nsq"
	"github.com/puoklam/chat-app-backend/api"
	"github.com/puoklam/chat-app-backend/db/model"
	"github.com/puoklam/chat-app-backend/middleware"
	"github.com/puoklam/chat-app-backend/mq"
	. "github.com/puoklam/chat-app-backend/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handlers struct {
	logger *log.Logger
}

func (h *Handlers) serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Println(err)
		return
	}
	u := r.Context().Value("user").(*model.User)
	s := r.Context().Value("session").(*model.Session)
	c := NewClient(&ClientCfg{
		Logger:    h.logger,
		Conn:      conn,
		User:      u,
		IP:        s.IP,
		Consumers: make(map[string][]*nsq.Consumer),
		Send:      make(chan Message, 256),
	})

	for _, g := range u.Groups {
		topic := g.Topic.String()
		ch := s.Ch

		// no need hash, should get session from context

		consumer, _ := mq.NewConsumer(topic, ch)
		consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
			var data mq.Message
			if err := json.Unmarshal(message.Body, &data); err != nil {
				return err
			}
			m := api.OutMessage{
				From: &api.OutUser{
					Base:        data.From.Base,
					Username:    data.From.Username,
					Displayname: data.From.Displayname,
				},
				Dst:       g.ID,
				DstType:   "group",
				Content:   string(data.Body),
				Timestamp: message.Timestamp,
			}
			b, err := json.Marshal(m)
			if err != nil {
				return err
			}
			msg := api.NewMessage(message, h.logger, b)
			c.Send() <- msg
			return nil
		}))
		if err = consumer.ConnectToNSQLookupd(os.Getenv("NSQLOOKUPD_ADDR")); err != nil {
			h.logger.Println(err)
			return
		}
		c.AddConsumer(topic, consumer)
	}

	GetHub().Register() <- c
	go c.WritePump()
	go c.ReadPump()
}

func (h *Handlers) connect(w http.ResponseWriter, r *http.Request) {
	h.serveWs(w, r)
}

func (h *Handlers) SetupRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticator(h.logger))
		r.Get("/", h.connect)
	})
}

func NewHandlers(logger *log.Logger) *Handlers {
	return &Handlers{logger}
}
