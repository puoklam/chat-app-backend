package auth

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
)

var (
	hs256Secret   any
	ed25519Secret any
)

func init() {
	hs256Secret = []byte(os.Getenv("HS256_SECRET"))
	data, _ := os.ReadFile("./ed25519")
	block, _ := pem.Decode(data)
	key, _ := x509.ParsePKCS8PrivateKey(block.Bytes)
	ed25519Secret = key
}

type Handlers struct {
	logger *log.Logger
}

func (h *Handlers) signin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SessionId *string `json:"sessionId"`
		Name      *string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if body.SessionId == nil || body.Name == nil {
		h.logger.Println("error: invalid format")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idToken, err := genIdToken(map[string]string{
		"id":   *body.Name,
		"name": *body.Name,
	})
	if err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	accessToken, err := genAccessToken(*body.SessionId, *body.Name)
	if err != nil {
		h.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "accessToken",
		Value:   accessToken,
		Expires: time.Now().Add(2 * time.Hour),
		// Domain:   "localhost",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "refreshToken",
		Value:   "refreshToken",
		Expires: time.Now().Add(60 * 24 * time.Hour),
		// Domain:   "localhost",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	})
	json.NewEncoder(w).Encode(struct {
		IdToken string `json:"idToken"`
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
}

func (h *Handlers) jwks(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./auth/jwks.json")
}

func (h *Handlers) SetupRoutes(r *chi.Mux) {
	r.Route("/auth", func(r chi.Router) {
		r.Get("/jwks.json", h.jwks)
		r.Post("/signin", h.signin)
		r.Post("/signout", h.signout)
	})
}

func NewHandlers(logger *log.Logger) *Handlers {
	return &Handlers{logger}
}
