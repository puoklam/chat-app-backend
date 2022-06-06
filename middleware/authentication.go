package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"github.com/puoklam/chat-app-backend/model"
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
			if claims, ok := t.Claims.(jwt.MapClaims); !ok || !t.Valid {
				w.WriteHeader(http.StatusUnauthorized)
				return
			} else {
				uid, sid := claims["sub"].(string), claims["aud"].(string)
				ctx := context.WithValue(context.WithValue(r.Context(), "user", &model.User{
					UserId: uid,
				}), "sessionId", sid)
				h.ServeHTTP(w, r.WithContext(ctx))
			}
		}
		return http.HandlerFunc(fn)
	}
}