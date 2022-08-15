package user

import (
	// "encoding/json"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/puoklam/chat-app-backend/db"
	"github.com/puoklam/chat-app-backend/db/model"
	"github.com/puoklam/chat-app-backend/middleware"
	"github.com/puoklam/chat-app-backend/mq"
	"github.com/puoklam/chat-app-backend/notifications"
	"github.com/puoklam/chat-app-backend/storage"
)

type Handlers struct {
	logger *log.Logger
}

func (h *Handlers) listUsers(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("s")
	users := make([]model.User, 0)
	db := db.GetDB(r.Context())
	if err := db.Model(&model.User{}).Where("displayname ILIKE ?", fmt.Sprintf("%%%s%%", s)).Find(&users).Error; err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(users)
}

func (h *Handlers) getUser(w http.ResponseWriter, r *http.Request) {
	oppo := r.Context().Value("opponent").(*model.User)
	encoder := json.NewEncoder(w)
	encoder.Encode(oppo)
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) updateUser(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user").(*model.User)
	if u.ID != r.Context().Value("opponent").(*model.User).ID {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<15)
	var body struct {
		Displayname *string `json:"displayname"`
		Bio         *string `json:"bio"`
	}
	encoder, decoder := json.NewEncoder(w), json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if body.Displayname == nil || *body.Displayname == "" || body.Bio == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing field"))
		return
	}
	if err := db.GetDB(r.Context()).Model(u).Select("Displayname", "Bio", "UpdatedAt").Updates(&model.User{Displayname: *body.Displayname, Bio: *body.Bio}).Error; err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := encoder.Encode(u); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Handlers) updateUserImage(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user").(*model.User)
	if u.ID != r.Context().Value("opponent").(*model.User).ID {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var limit int64 = 1 << 19 // 0.5 MB
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	if err := r.ParseMultipartForm(limit); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("image too large"))
		return
	}
	image, fh, _ := r.FormFile("image")
	if image != nil {
		defer image.Close()
	}
	if image == nil || fh == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no image"))
		return
	}
	if fh.Size > 4<<20 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("image too large"))
		return
	}
	p := make([]byte, 512)
	if _, err := image.Read(p); err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}
	if !validTypes[http.DetectContentType(p)] {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unsupported image format"))
		return
	}
	buf := bytes.NewBuffer(nil)
	image.Seek(0, 0)
	if _, err := io.Copy(buf, image); err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	url, err := storage.Upload(r.Context(), buf, "groups/")
	if err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := db.GetDB(r.Context()).Model(u).Select("ImageURL").Updates(&model.User{ImageURL: url}).Error; err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(u); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handlers) updateNotifications(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user").(*model.User)
	rel := r.Context().Value("rel").(*model.Relationship)
	if rel == nil {
		w.WriteHeader(http.StatusUnauthorized)
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
	if rel.User1ID == u.ID {
		rel.ForwardNotifications = *body.Notifications
	} else {
		rel.BackwardNotifications = *body.Notifications
	}
	if err := db.GetDB(r.Context()).Save(rel).Error; err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) createMsg(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user").(*model.User)
	rel := r.Context().Value("rel").(*model.Relationship)
	if rel == nil || rel.ForwardStatus != model.StatusAccepted || rel.BackwardStatus != model.StatusAccepted {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var body struct {
		Message *string `json:"message"`
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil || body.Message == nil || *body.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	msg := &mq.BroadCastMessage{
		From: mq.User{
			Base:        u.Base,
			Username:    u.Username,
			Displayname: u.Displayname,
			ImageURL:    u.ImageURL,
		},
		Body: []byte(*body.Message),
	}
	if err := mq.Publish(rel.Host, rel.Topic.String(), msg); err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// TODO: handle error
	go h.send(rel, u.Displayname, *body.Message)
	w.WriteHeader(http.StatusOK)
	w.WriteHeader(http.StatusCreated)
}

func (h *Handlers) send(rel *model.Relationship, title, message string) error {
	userIds := make([]uint, 0)

	if rel.ForwardNotifications {
		userIds = append(userIds, rel.User1ID)
	}
	if rel.BackwardNotifications {
		userIds = append(userIds, rel.User2ID)
	}
	tokens := make([]string, 0)
	if err := db.GetDB(context.Background()).Model(&model.Session{}).Select("expo_push_token").Where("user_id IN ? AND status = ? AND expo_push_token != ?", userIds, model.StatusOffline, "").Find(&tokens).Error; err != nil {
		h.logger.Println(err)
		return err
	}
	return notifications.Send(title, message, tokens)
}

func (h *Handlers) SetupRoutes(r *chi.Mux) {
	r.Route("/users", func(r chi.Router) {
		r.Use(middleware.Authenticator(h.logger))
		r.Get("/", h.listUsers)
		r.Group(func(r chi.Router) {
			r.Use(middleware.WithOpponent)
			r.Get("/{userID}", h.getUser)
			r.Patch("/{userID}", h.updateUser)
			r.Post("/{userID}/image", h.updateUserImage)
			r.With(middleware.WithRel).Post("/{userID}/messages", h.createMsg)
			r.With(middleware.WithRel).Post("/{userID}/notifications", h.updateNotifications)
		})
	})
}

func NewHandlers(l *log.Logger) *Handlers {
	return &Handlers{l}
}
