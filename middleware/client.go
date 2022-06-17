package middleware

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
