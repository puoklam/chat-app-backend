package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/puoklam/chat-app-backend/db"
	"github.com/puoklam/chat-app-backend/db/model"
	"github.com/puoklam/chat-app-backend/env"
	"github.com/puoklam/chat-app-backend/middleware"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handlers struct {
	logger *log.Logger
}

func (h *Handlers) signin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username *string `json:"username"`
		Password *string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if body.Username == nil || body.Password == nil {
		h.logger.Println("error: invalid format")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c := r.Context()
	u, err := getUserFromUsername(c, *body.Username)
	if err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if u == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Pass), []byte(*body.Password)) != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ip := c.Value("deviceIP").(string)
	s := &model.Session{}
	if err := db.GetDB(c).Where(&model.Session{UserID: u.ID, IP: ip}).First(s).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// insert new session record
		if s, err = insertSession(c, u.ID, ip, c.Value("expoPushToken").(string)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	// multiple users in single device
	// if s.UserID != u.ID {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	w.Write([]byte("multiple users"))
	// 	return
	// }

	idToken, err := genIdToken(map[string]any{
		"id":          u.ID,
		"username":    u.Username,
		"displayname": u.Displayname,
	})
	if err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	accessToken, err := genAccessToken(r.Context().Value("deviceIP").(string), strconv.FormatUint(uint64(u.ID), 10))
	if err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		Expires:  time.Now().Add(2 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		// MaxAge:   int(7200),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    "refreshToken",
		Expires:  time.Now().Add(60 * 24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		// MaxAge:   int(60 * 24 * 60),
	})
	json.NewEncoder(w).Encode(struct {
		IdToken string `json:"id_token"`
	}{
		IdToken: idToken,
	})
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) signout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username *string `json:"username"`
		Password *string `json:"password"`
	}
	encoder, decoder := json.NewEncoder(w), json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if body.Username == nil || body.Password == nil || err != nil {
		if err != nil {
			h.logger.Println(err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(*body.Username) == 0 || len(*body.Password) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if user, err := getUserFromUsername(r.Context(), *body.Username); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if user != nil {
		w.WriteHeader(http.StatusConflict)
		encoder.Encode("username exists")
	}
	db := db.GetDB(r.Context())
	passBytes, err := bcrypt.GenerateFromPassword([]byte(*body.Password), 14)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user := &model.User{
		Username: *body.Username,
		Pass:     string(passBytes),
	}
	if db.Create(user).Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	encoder.Encode(user)
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) jwks(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, env.JWKS_PATH)
}

func (h *Handlers) user(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user").(*model.User)

	// db := db.GetDB(r.Context())
	// var a any
	// db.Model(u).Preload("Memberships.Group").Limit(1).Find(u)
	// h.logger.Printf("%+v", u)
	encoder := json.NewEncoder(w)

	w.WriteHeader(http.StatusOK)
	if err := encoder.Encode(u); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	return
}

func (h *Handlers) SetupRoutes(r *chi.Mux) {
	r.Route("/auth", func(r chi.Router) {
		r.Get("/jwks.json", h.jwks)
		r.Post("/register", h.register)
		r.With(middleware.WithExpoPushToken).Post("/signin", h.signin)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticator(h.logger))
			r.With(middleware.NoCache).Get("/user", h.user)
			r.Post("/signout", h.signout)
		})
	})
}

func getUserFromUsername(ctx context.Context, un string) (user *model.User, err error) {
	user = &model.User{}
	if ctx == nil {
		ctx = context.Background()
	}
	if err = db.GetDB(ctx).First(user, "username = ?", un).Error; err != nil {
		user = nil
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		}
	}
	return
}

func insertSession(ctx context.Context, userID uint, ip string, token string) (session *model.Session, err error) {
	k := fmt.Sprintf("%s:%s", strconv.FormatUint(uint64(userID), 10), ip)

	h := sha256.New()
	h.Write([]byte(k))
	ch := hex.EncodeToString(h.Sum(nil))

	session = &model.Session{
		UserID:        userID,
		IP:            ip,
		Ch:            ch,
		ExpoPushToken: token,
		Status:        model.StatusOffline,
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err = db.GetDB(ctx).Create(session).Error; err != nil {
		session = nil
	}
	return
}

func NewHandlers(logger *log.Logger) *Handlers {
	return &Handlers{logger}
}
