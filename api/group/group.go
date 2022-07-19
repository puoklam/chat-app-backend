package group

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nsqio/go-nsq"
	"github.com/puoklam/chat-app-backend/db"
	"github.com/puoklam/chat-app-backend/db/model"
	"github.com/puoklam/chat-app-backend/env"
	"github.com/puoklam/chat-app-backend/middleware"
	"github.com/puoklam/chat-app-backend/mq"
	"gorm.io/gorm"
)

type Handlers struct {
	logger *log.Logger
}

func (h *Handlers) listGroups(w http.ResponseWriter, r *http.Request) {

	grps := make([]OutListGroups, 0)
	db := db.GetDB(r.Context())
	if err := db.Model(&model.Group{}).Find(&grps).Error; err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(grps)
}

func (h *Handlers) getGroup(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user").(*model.User)
	g := r.Context().Value("group").(*model.Group)

	q := "SELECT EXISTS(SELECT 1 FROM memberships WHERE user_id = ? AND group_id = ?)"
	var exists bool
	if err := db.GetDB(r.Context()).Raw(q, u.ID, g.ID).Scan(&exists).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(&OutGetGroup{
		Base:   g.Base,
		Name:   g.Name,
		Joined: exists,
	})
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
		Name: *body.Name,
		Host: env.NSQD_TCP_ADDR,
		Memberships: []*model.Membership{{
			UserID: u.ID,
			Status: model.StatusActive,
		}},
		// Members: []*model.User{u},
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(g).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&model.Session{}).Where(&model.Session{Status: model.StatusOnline}).Count(&count).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Conn{
			UserID: u.ID,
			Topic:  g.Topic.String(),
			Count:  0,
		}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	topic := g.Topic.String()

	// Init nsq topic
	msg := &mq.BroadCastMessage{
		From: mq.User{
			Base:        u.Base,
			Username:    u.Username,
			Displayname: u.Displayname,
		},
		Body: []byte(fmt.Sprintf("group %s created", g.Name)),
	}

	h.postJoin(r, g)

	if err := mq.Publish(g.Host, topic, msg); err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if exists {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("already joined"))
		return
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		// if err := db.Model(g).Association("Members").Append(u); err != nil {
		if err := tx.Create(&model.Membership{UserID: u.ID, GroupID: g.ID, Status: model.StatusActive}).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&model.Session{}).Where(&model.Session{Status: model.StatusOnline}).Count(&count).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Conn{
			UserID: u.ID,
			Topic:  g.Topic.String(),
			Count:  0,
		}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	topic := g.Topic.String()

	h.postJoin(r, g)

	// Send greeting msg
	msg := &mq.BroadCastMessage{
		From: mq.User{
			Base:        u.Base,
			Username:    u.Username,
			Displayname: u.Displayname,
		},
		Body: []byte(fmt.Sprintf("%s joined the group", u.Displayname)),
	}
	if err := mq.Publish(g.Host, topic, msg); err != nil {
		h.logger.Println(err)
	}

	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(&OutGetGroup{
		Base:   g.Base,
		Name:   g.Name,
		Joined: true,
	})
}

func (h *Handlers) exitGroup(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user").(*model.User)
	g := r.Context().Value("group").(*model.Group)

	db := db.GetDB(r.Context())

	m := &model.Membership{}
	if err := db.Where(&model.Membership{UserID: u.ID, GroupID: g.ID}).First(m).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
		return
	}

	m.Status = model.StatusDeleting
	if err := db.Save(m).Error; err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		// if err := tx.Model(g).Association("Members").Delete(u); err != nil {
		// 	return err
		// }
		if err := tx.Delete(m).Error; err != nil {
			return err
		}
		if err := h.postExit(r, g); err != nil {
			return err
		}
		if err := tx.Delete(&model.Conn{UserID: u.ID, Topic: g.Topic.String()}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		m.Status = model.StatusActive
		// TODO: handle error
		db.Save(m)
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) postJoin(r *http.Request, g *model.Group) {
	u := r.Context().Value("user").(*model.User)
	topic := g.Topic.String()

	// WriteCommand not working for multiple subscribes, instantiate consumers instead
	// TODO: log error
	for _, s := range u.Sessions {
		cfg := nsq.NewConfig()
		delegate := &mq.ConnDelegate{}
		conn := nsq.NewConn(g.Host, cfg, delegate)
		if _, err := conn.Connect(); err != nil {
			h.logger.Println(err)
			continue
		}
		cmd := nsq.Subscribe(topic, s.Ch)
		if err := conn.WriteCommand(cmd); err != nil {
			h.logger.Println(err)
		}
		conn.Close()
	}
	msg := &mq.ExchangeMessage{
		Type:    mq.SignalAddConsumers,
		UserID:  u.ID,
		GroupID: g.ID,
		Topic:   topic,
	}
	mq.Publish(env.EXCHANGE_NSQD_TCP_ADDR, "info", msg)
}

func (h *Handlers) postExit(r *http.Request, g *model.Group) error {
	u := r.Context().Value("user").(*model.User)

	topic := g.Topic.String()

	msg := &mq.ExchangeMessage{
		Type:          mq.SignalClearConsumers,
		UserID:        u.ID,
		GroupID:       g.ID,
		Topic:         topic,
		PostbackTopic: env.SERVER_ID,
		PostbackCh:    env.SERVER_ID,
		PostbackMsg: &mq.PrivateMessage{
			Type:   mq.SignalClearConsumersComplete,
			UserID: u.ID,
			Topic:  topic,
		},
	}
	if mq.CleanUpChans[u.ID] == nil {
		mq.CleanUpChans[u.ID] = make(map[string]chan bool)
	}
	cleanUpCh := make(chan bool, 1)
	mq.CleanUpChans[u.ID][topic] = cleanUpCh
	mq.Publish(env.EXCHANGE_NSQD_TCP_ADDR, "info", msg)
	// TODO: select with ch and timeout, recevie either break loop
	res := <-cleanUpCh
	if res {
		// successfully cleanup consummers with given user id and topic
		// start unsubscribe
		for _, s := range u.Sessions {
			// Unsubscribe command doesn't work, directly fetch instead
			// url := fmt.Sprintf("http://%s/channel/delete?topic=%s&channel=%s", g.Host, topic, s.Ch)
			url := fmt.Sprintf("%s/channel/delete?topic=%s&channel=%s", env.NSQD_API_ADDR, topic, s.Ch)
			_, err := http.Post(url, "application/json", nil)
			if err != nil {
				return err
			}
		}
	} else {
		// TODO: channel recevie false
		return fmt.Errorf("error when distributing clear consumers signal")
	}
	return nil
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
	msg := &mq.BroadCastMessage{
		From: mq.User{
			Base:        u.Base,
			Username:    u.Username,
			Displayname: u.Displayname,
		},
		Body: []byte(*body.Message),
	}
	if err := mq.Publish(g.Host, g.Topic.String(), msg); err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// notifications.Send(r.Context(), map[string]string{
	// 	"a": "test",
	// 	"b": "test2",
	// }, "")
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) SetupRoutes(r *chi.Mux) {
	r.Route("/groups", func(r chi.Router) {
		r.Use(middleware.Authenticator(h.logger))
		r.Get("/", h.listGroups)
		r.Post("/", h.createGroup)
		r.With(middleware.WithGroup).Get("/{groupID}", h.getGroup)
		r.With(middleware.WithGroup).Post("/{groupID}/join", h.joinGroup)
		r.With(middleware.WithGroup).Post("/{groupID}/exit", h.exitGroup)
		r.With(middleware.WithGroup).Post("/{groupID}/messages", h.createMsg)
	})
}

func NewHandlers(l *log.Logger) *Handlers {
	return &Handlers{l}
}
