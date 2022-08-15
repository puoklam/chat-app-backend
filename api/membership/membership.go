package membership

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/puoklam/chat-app-backend/db"
	"github.com/puoklam/chat-app-backend/db/model"
	"github.com/puoklam/chat-app-backend/middleware"
	"gorm.io/gorm"
)

type Handlers struct {
	logger *log.Logger
}

func (h *Handlers) updateNotifications(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user").(*model.User)
	g := r.Context().Value("group").(*model.Group)

	m := &model.Membership{}
	if err := db.GetDB(r.Context()).Where(&model.Membership{UserID: u.ID, GroupID: g.ID}).First(m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	var body struct {
		Notifications *bool `json:"notifications"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if body.Notifications == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing field: notifications"))
		return
	}
	m.Notification = *body.Notifications
	if err := db.GetDB(r.Context()).Save(m).Error; err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) SetupRoutes(r *chi.Mux) {
	r.Route("/memberships", func(r chi.Router) {
		r.Use(middleware.Authenticator(h.logger))
		r.With(middleware.WithGroup).Post("/{groupID}/notifications", h.updateNotifications)
	})
}

func NewHandlers(l *log.Logger) *Handlers {
	return &Handlers{l}
}
