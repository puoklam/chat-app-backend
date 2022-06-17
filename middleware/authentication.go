package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"github.com/puoklam/chat-app-backend/db"
	"github.com/puoklam/chat-app-backend/db/model"
	"gorm.io/gorm"
)

var hs256Secret any

func init() {
	hs256Secret = []byte(os.Getenv("HS256_SECRET"))
}

func Authenticator(logger *log.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie("accessToken")
			if err != nil {
				logger.Println(err)
				if errors.Is(err, http.ErrNoCookie) {
					w.WriteHeader(http.StatusUnauthorized)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
				return
			}
			// verify jwt
			t, err := jwt.Parse(c.Value, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
				}
				return hs256Secret, nil
			})
			// match session
			if claims, ok := t.Claims.(jwt.MapClaims); !ok || !t.Valid || claims["aud"] != r.Context().Value("deviceIP") {
				w.WriteHeader(http.StatusUnauthorized)
				return
			} else {
				uid := claims["sub"].(string)
				ip := claims["aud"].(string)
				db := db.GetDB(r.Context())
				var u model.User
				if err := db.Preload("Groups").Preload("Sessions").First(&u, uid).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						w.WriteHeader(http.StatusForbidden)
					} else {
						w.WriteHeader(http.StatusInternalServerError)
					}
					return
				}
				var s *model.Session
				for _, ss := range u.Sessions {
					if ss.IP == ip {
						s = &ss
						break
					}
				}
				if s == nil {
					w.WriteHeader(http.StatusForbidden)
					w.Write([]byte("session does not exist"))
					return
				}
				ctx := context.WithValue(context.WithValue(r.Context(), "user", &u), "session", s)
				h.ServeHTTP(w, r.WithContext(ctx))
			}
		}
		return http.HandlerFunc(fn)
	}
}
