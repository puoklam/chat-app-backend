package ws

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/nsqio/go-nsq"
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
		logger:    h.logger,
		hub:       hub,
		conn:      conn,
		producer:  mq.GetProducer(),
		consumers: make(map[string]*nsq.Consumer),
		send:      make(chan []byte, 256),
		receive:   make(chan []byte, 1024),
	}
	consumer, err := mq.NewConsumer("test", "bar")
	if err != nil {
		h.logger.Println(err)
		return
	}
	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		c.send <- message.Body
		return nil
	}))
	if err = consumer.ConnectToNSQLookupd("127.0.0.1:4161"); err != nil {
		h.logger.Println(err)
		return
	}
	c.AddConsumer("test", consumer)
	c.hub.register <- c
	go c.writePump()
	go c.ReadPump()
}

func (h *Handlers) connect(w http.ResponseWriter, r *http.Request) {
	hub := GetHub()
	h.serveWs(hub, w, r)
}

func (h *Handlers) SetupRoutes(r *chi.Mux) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello, world! haha"))
	})
	r.Route("/ws", func(r chi.Router) {
		r.Get("/", h.connect)
	})
}

func NewHandlers(logger *log.Logger) *Handlers {
	return &Handlers{logger}
}
