package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	. "github.com/puoklam/chat-app-backend/db"
	"github.com/puoklam/chat-app-backend/db/model"
	"gorm.io/gorm"
)

func WithGroup(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "groupID")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		db := GetDB(r.Context())
		var grp model.Group
		if err := db.Preload("Memberships").First(&grp, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		ctx := context.WithValue(r.Context(), "group", &grp)
		h.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
