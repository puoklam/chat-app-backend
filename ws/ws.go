package ws

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/nsqio/go-nsq"
	"github.com/puoklam/chat-app-backend/middleware"
	"github.com/puoklam/chat-app-backend/model"

	// "github.com/puoklam/chat-app-backend/model"
	"github.com/puoklam/chat-app-backend/mq"
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

func (h *Handlers) serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Println(err)
		return
	}
	c := &Client{
		logger: h.logger,
		hub:    hub,
		conn:   conn,
		// userId:    r.Context().Value("user").(*model.User).UserId,
		producer:  mq.GetProducer(),
		consumers: make(map[string]*nsq.Consumer),
		send:      make(chan *nsq.Message, 256),
	}

	userId := r.Context().Value("user").(*model.User).UserId
	// TODO: find group chats and conversation from db
	consumer1, _ := mq.NewConsumer("group1", userId)
	consumer2, _ := mq.NewConsumer("group2", userId)

	consumer1.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		c.send <- message
		return nil
	}))
	consumer2.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		c.send <- message
		return nil
	}))
	if err = consumer1.ConnectToNSQLookupd(os.Getenv("NSQLOOKUPD_ADDR")); err != nil {
		h.logger.Println(err)
		return
	}
	if err = consumer2.ConnectToNSQLookupd(os.Getenv("NSQLOOKUPD_ADDR")); err != nil {
		h.logger.Println(err)
		return
	}
	c.AddConsumer("group1", consumer1)
	c.AddConsumer("group2", consumer2)
	c.hub.register <- c
	go c.writePump()
	go c.ReadPump()
}

func (h *Handlers) connect(w http.ResponseWriter, r *http.Request) {
	hub := GetHub()
	h.serveWs(hub, w, r)
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
