package middleware

import "net/http"

func NoCache(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate;")
		w.Header().Set("pragma", "no-cache")
		// w.Header().Set("X-Content-Type-Options", "nosniff")
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
