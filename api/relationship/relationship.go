package relationship

import (
	"encoding/json"
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

func (h *Handlers) getRelationship(w http.ResponseWriter, r *http.Request) {
	rel := r.Context().Value("rel").(*model.Relationship)
	if rel == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(rel)
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) updateRelationship(w http.ResponseWriter, r *http.Request) {
	u1 := r.Context().Value("user").(*model.User)
	u2 := r.Context().Value("opponent").(*model.User)
	rel := r.Context().Value("rel").(*model.Relationship)

	decoder := json.NewDecoder(r.Body)
	var body struct {
		Status *string `json:"status"`
	}
	err := decoder.Decode(&body)
	if body.Status == nil || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	db := db.GetDB(r.Context())
	if rel == nil {
		rel = &model.Relationship{
			User1ID:               u1.ID,
			User2ID:               u2.ID,
			ForwardStatus:         model.StatusAccepted,
			BackwardStatus:        model.StatusDefault,
			Host:                  env.NSQD_TCP_ADDR,
			ForwardNotifications:  true,
			BackwardNotifications: true,
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&rel).Error; err != nil {
				return err
			}
			if err := tx.Create(&model.Conn{
				UserID: u1.ID,
				Topic:  rel.Topic.String(),
				Count:  0,
			}).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			h.postUpdate(r, rel)
			w.WriteHeader(http.StatusOK)
		}
		return
	}
	var init bool
	if rel.User1ID == u1.ID {
		rel.ForwardStatus = *body.Status
	} else {
		if rel.BackwardStatus == model.StatusDefault {
			init = true
		}
		rel.BackwardStatus = *body.Status
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(rel).Error; err != nil {
			return err
		}
		if !init {
			return nil
		}
		if err := tx.Create(&model.Conn{
			UserID: u1.ID,
			Topic:  rel.Topic.String(),
			Count:  0,
		}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if init {
		h.postUpdate(r, rel)
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) postUpdate(r *http.Request, rel *model.Relationship) {
	u := r.Context().Value("user").(*model.User)
	topic := rel.Topic.String()
	for _, s := range u.Sessions {
		cfg := nsq.NewConfig()
		delegate := &mq.ConnDelegate{}
		conn := nsq.NewConn(rel.Host, cfg, delegate)
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
	targetID := rel.User2ID
	if u.ID == targetID {
		targetID = rel.User1ID
	}
	msg := &mq.ExchangeMessage{
		Type:       mq.SignalAddConsumers,
		UserID:     u.ID,
		TargetID:   targetID,
		TargetType: "personal",
		Topic:      topic,
	}
	mq.Publish(env.EXCHANGE_NSQD_TCP_ADDR, "info", msg)
}

func (h *Handlers) SetupRoutes(r *chi.Mux) {
	r.Route("/relationships", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticator(h.logger), middleware.WithOpponent, middleware.WithRel)
			r.Get("/{userID}", h.getRelationship)
			r.Post("/{userID}", h.updateRelationship)
		})
	})
}

func NewHandlers(l *log.Logger) *Handlers {
	return &Handlers{l}
}
