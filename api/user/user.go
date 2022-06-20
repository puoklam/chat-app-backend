package user

import (
	// "encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	// "github.com/puoklam/chat-app-backend/db/model"
	"github.com/puoklam/chat-app-backend/middleware"
)

type Handlers struct {
	logger *log.Logger
}

// func (h *Handlers) me(w http.ResponseWriter, r *http.Request) {
// 	u := r.Context().Value("user").(*model.User)
// 	encode := json.NewEncoder(w)

// 	w.WriteHeader(http.StatusOK)
// 	if err := encode.Encode(u); err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 	}
// 	return
// }

func (h *Handlers) getUser(w http.ResponseWriter, r *http.Request) {

}

func (h *Handlers) SetupRoutes(r *chi.Mux) {
	r.Route("/users", func(r chi.Router) {
		r.Use(middleware.Authenticator(h.logger))
		// r.Get("/me", h.me)
		r.Get("/{userID}", h.getUser)
	})
}

func NewHandlers(l *log.Logger) *Handlers {
	return &Handlers{l}
}
