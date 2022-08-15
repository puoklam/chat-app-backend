package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/puoklam/chat-app-backend/db"
	"github.com/puoklam/chat-app-backend/db/model"
	"gorm.io/gorm"
)

func WithOpponent(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "userID")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var u model.User
		if err := db.GetDB(r.Context()).First(&u, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		ctx := context.WithValue(r.Context(), "opponent", &u)
		h.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// must use after withOpponent
func WithRel(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		u1 := r.Context().Value("user").(*model.User)
		u2 := r.Context().Value("opponent").(*model.User)
		if u1.ID == u2.ID {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var rel model.Relationship
		err := db.GetDB(r.Context()).
			Where(&model.Relationship{User1ID: u1.ID, User2ID: u2.ID}).
			Or(&model.Relationship{User1ID: u2.ID, User2ID: u1.ID}).
			Preload("User1").
			Preload("User2").
			First(&rel).
			Error
		var ctx context.Context
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			ctx = context.WithValue(r.Context(), "rel", (*model.Relationship)(nil))
		} else {
			ctx = context.WithValue(r.Context(), "rel", &rel)
		}
		h.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
