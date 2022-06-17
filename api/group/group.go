package group

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/nsqio/go-nsq"
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
	if err := mq.GetProducer().Publish(topic, []byte(fmt.Sprintf("group %s created", g.Name))); err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.postJoin(w, r, topic)

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

	h.postJoin(w, r, topic)

	if err := mq.GetProducer().Publish(topic, []byte(fmt.Sprintf("%s joined the group", u.Displayname))); err != nil {
		h.logger.Println(err)
	}

	w.WriteHeader(http.StatusOK)
	return
}

func (h *Handlers) postJoin(w http.ResponseWriter, r *http.Request, topic string) {
	// conn := mq.GetConn()
	// for _, s := range u.Sessions {
	// 	cmd := nsq.Subscribe(topic, s.Ch)
	// 	if err := conn.WriteCommand(cmd); err != nil {
	// 		h.logger.Println(err)
	// 	}
	// }

	// WriteCommand not working for multiple subscribes, instantiate consumers instead
	// TODO: log error
	u := r.Context().Value("user").(*model.User)
	for _, s := range u.Sessions {
		c := ws.GetHub().Client(u.ID, s.IP)
		consumer, err := mq.NewConsumer(topic, s.Ch)
		if err != nil {
			continue
		}
		consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
			if c == nil {
				message.RequeueWithoutBackoff(0)
			} else {
				c.Send() <- message
			}
			return nil
		}))
		if consumer.ConnectToNSQLookupd(os.Getenv("NSQLOOKUPD_ADDR")) != nil || c == nil {
			// client not connected
			consumer.Stop()
			continue
		}
		c.AddConsumer(topic, consumer)
	}
}

func (h *Handlers) SetupRoutes(r *chi.Mux) {
	r.Route("/groups", func(r chi.Router) {
		r.Use(middleware.Authenticator(h.logger))
		r.Get("/", h.listGroups)
		r.Post("/", h.createGroup)
		r.With(middleware.WithGroup).Post("/{groupID}/join", h.joinGroup)
	})
}

func NewHandlers(l *log.Logger) *Handlers {
	return &Handlers{l}
}
