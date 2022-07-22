package socket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/nsqio/go-nsq"
	"github.com/puoklam/chat-app-backend/api"
	"github.com/puoklam/chat-app-backend/db/model"
	"github.com/puoklam/chat-app-backend/env"
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
		Logger:  h.logger,
		Conn:    conn,
		Session: s,
		User:    u,
		// IP:        s.IP,
		Consumers: make(map[string][]*nsq.Consumer),
		Send:      make(chan Message, 256),
	})

	// db := db.GetDB(r.Context())
	// groups := make([]model.Group, 0)
	// if err := db.Model(&model.Membership{}).Where(&model.Membership{UserID: u.ID, Status: model.StatusActive}).Select("groups.*").Joins("LEFT JOIN groups ON group_id = groups.id", u.ID, model.StatusDeleting).Scan(&groups).Error; err != nil {
	// 	h.logger.Println(err)
	// 	return
	// }
	for _, m := range u.Memberships {
		g := m.Group
		topic := g.Topic.String()
		ch := s.Ch
		gid := g.ID
		consumer, _ := mq.NewConsumer(topic, ch)
		consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
			var data mq.BroadCastMessage
			if err := json.Unmarshal(message.Body, &data); err != nil {
				return err
			}
			m := api.OutMessage{
				// From: &api.OutUser{
				// 	Base:        data.From.Base,
				// 	Username:    data.From.Username,
				// 	Displayname: data.From.Displayname,
				// },
				FromID:       data.From.ID,
				FromName:     data.From.Displayname,
				FromImageURL: data.From.ImageURL,
				Dst:          gid,
				DstType:      "group",
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
		if err = consumer.ConnectToNSQLookupd(env.NSQLOOKUPD_ADDR); err != nil {
			h.logger.Println(err)
			return
		}
		if err := c.AddConsumer(r.Context(), topic, consumer); err != nil {
			consumer.Stop()
		}
	}

	if err := c.StartWithContext(r.Context()); err != nil {
		c.ClearConsumers()
		h.logger.Println(err)
		return
	}
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
