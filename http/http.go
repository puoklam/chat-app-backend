package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/puoklam/chat-app-backend/middleware"
	"github.com/puoklam/chat-app-backend/mq"
)

type Body struct {
	Topic string `json:"topic"`
}

type Handlers struct {
	logger *log.Logger
}

func (h *Handlers) newTopic(w http.ResponseWriter, r *http.Request) {
	var body Body
	json.NewDecoder(r.Body).Decode(&body)
	mq.GetProducer().Publish(body.Topic, []byte("init"))
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) SetupRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticator(h.logger))
		r.Post("/topic", h.newTopic)
	})
}

func NewHandlers(logger *log.Logger) *Handlers {
	return &Handlers{logger}
}
