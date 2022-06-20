package group

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/nsqio/go-nsq"
	"github.com/puoklam/chat-app-backend/api"
	"github.com/puoklam/chat-app-backend/db"
	"github.com/puoklam/chat-app-backend/db/model"
	"github.com/puoklam/chat-app-backend/middleware"
	"github.com/puoklam/chat-app-backend/mq"
	"github.com/puoklam/chat-app-backend/ws"
)

type Handlers struct {
	logger *log.Logger
}

func (h *Handlers) listGroups(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)

	grps := make([]OutListGroups, 0)
	db := db.GetDB(r.Context())
	if err := db.Model(&model.Group{}).Find(&grps).Error; err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	encoder.Encode(grps)
}

func (h *Handlers) createGroup(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user").(*model.User)
	var body InCreateGroup
	encoder, decoder := json.NewEncoder(w), json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if body.Name == nil || *body.Name == "" || err != nil {
		if err != nil {
			h.logger.Println(err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create group record
	db := db.GetDB(r.Context())
	g := &model.Group{
		Name:    *body.Name,
		Members: []*model.User{r.Context().Value("user").(*model.User)},
	}
	if err := db.Create(g).Error; err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	topic := g.Topic.String()

	// Init nsq topic
	msg := &mq.Message{
		From: mq.User{
			Base:        u.Base,
			Username:    u.Username,
			Displayname: u.Displayname,
		},
		Body: []byte(fmt.Sprintf("group %s created", g.Name)),
	}
	if err := mq.Publish(topic, msg); err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.postJoin(w, r, g)

	w.WriteHeader(http.StatusOK)
	encoder.Encode(&OutCreateGroup{g.Base, g.Name})
}

func (h *Handlers) joinGroup(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user").(*model.User)
	g := r.Context().Value("group").(*model.Group)
	db := db.GetDB(r.Context())

	// TODO: check if record exists
	var exists bool
	if err := db.Raw("SELECT EXISTS(SELECT 1 FROM memberships WHERE user_id = ? AND group_id = ?)", u.ID, g.ID).Scan(&exists).Error; err != nil {
		h.logger.Println(err)
	}
	if exists {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("already joined"))
		return
	}

	if err := db.Model(g).Association("Members").Append(u); err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	topic := g.Topic.String()

	h.postJoin(w, r, g)

	// Send greeting msg
	msg := &mq.Message{
		From: mq.User{
			Base:        u.Base,
			Username:    u.Username,
			Displayname: u.Displayname,
		},
		Body: []byte(fmt.Sprintf("%s joined the group", u.Displayname)),
	}
	if err := mq.Publish(topic, msg); err != nil {
		h.logger.Println(err)
	}

	w.WriteHeader(http.StatusOK)
	return
}

func (h *Handlers) postJoin(w http.ResponseWriter, r *http.Request, g *model.Group) {
	u := r.Context().Value("user").(*model.User)
	topic := g.Topic.String()

	// WriteCommand not working for multiple subscribes, instantiate consumers instead
	// TODO: log error
	for _, s := range u.Sessions {
		conn := mq.GetConn()
		cfg := nsq.NewConfig()
		delegate := &mq.ConnDelegate{}
		conn = nsq.NewConn(os.Getenv("NSQD_ADDR"), cfg, delegate)
		if _, err := conn.Connect(); err != nil {
			log.Println(err)
			continue
		}
		cmd := nsq.Subscribe(topic, s.Ch)
		if err := conn.WriteCommand(cmd); err != nil {
			h.logger.Println(err)
		}
		conn.Close()

		c := ws.GetHub().Client(u.ID, s.IP)
		consumer, err := mq.NewConsumer(topic, s.Ch)
		if err != nil {
			continue
		}
		if c == nil {
			continue
		}
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
		if consumer.ConnectToNSQLookupd(os.Getenv("NSQLOOKUPD_ADDR")) != nil {
			consumer.Stop()
			continue
		}
		c.AddConsumer(topic, consumer)
	}
}

func (h *Handlers) createMsg(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user").(*model.User)
	g := r.Context().Value("group").(*model.Group)

	q := "SELECT EXISTS(SELECT 1 FROM memberships WHERE user_id = ? AND group_id = ?)"
	var exists bool
	if err := db.GetDB(r.Context()).Raw(q, u.ID, g.ID).Scan(&exists).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var body InCreateMsg
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if body.Message == nil || *body.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	msg := &mq.Message{
		From: mq.User{
			Base:        u.Base,
			Username:    u.Username,
			Displayname: u.Displayname,
		},
		Body: []byte(*body.Message),
	}
	if err := mq.Publish(g.Topic.String(), msg); err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) SetupRoutes(r *chi.Mux) {
	r.Route("/groups", func(r chi.Router) {
		r.Use(middleware.Authenticator(h.logger))
		r.Get("/", h.listGroups)
		r.Post("/", h.createGroup)
		r.With(middleware.WithGroup).Post("/{groupID}/join", h.joinGroup)
		r.With(middleware.WithGroup).Post("/{groupID}/messages", h.createMsg)
	})
}

func NewHandlers(l *log.Logger) *Handlers {
	return &Handlers{l}
}
