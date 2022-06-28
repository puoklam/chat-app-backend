package server

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/puoklam/chat-app-backend/env"
	m "github.com/puoklam/chat-app-backend/middleware"
)

func New(mux http.Handler) *http.Server {
	// see https://blog.cloudflare.com/exposing-go-on-the-internet/
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
		},
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
	}

	return &http.Server{
		Addr:              ":" + env.APP_PORT,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		TLSConfig:         tlsConfig,
		Handler:           mux,
	}
}

func SetupMiddlewares(r *chi.Mux) {
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"Origin", "Accept", "Content-Type"},
		AllowCredentials: true,
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(m.WithDeviceInfo)
	// r.Use(m.WithClient)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
}
