package middleware

import (
	"context"
	"net/http"
)

// import (
// 	"context"
// 	"net/http"

// 	"github.com/puoklam/chat-app-backend/ws"
// )

// func WithClient(h http.Handler) http.Handler {
// 	fn := func(w http.ResponseWriter, r *http.Request) {
// 		hub := ws.GetHub()
// 		ip := r.Context().Value("deviceIP").(string)
// 		c := hub.Client()
// 		ctx := context.WithValue(r.Context(), "client", c)
// 		h.ServeHTTP(w, r.WithContext(ctx))
// 	}
// 	return http.HandlerFunc(fn)
// }

func WithExpoPushToken(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t := r.Header.Get("X-Expo-Push-Token")
		if t == "" {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("missing header: X-Expo-Push-Token"))
			return
		}
		ctx := context.WithValue(r.Context(), "expoPushToken", t)
		h.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
