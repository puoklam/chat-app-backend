package middleware

import (
	"context"
	"net/http"
	"strings"
)

func WithDeviceInfo(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ip := strings.Split(r.RemoteAddr, ":")[0]
		ctx := context.WithValue(r.Context(), "deviceIP", ip)
		h.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
