package mux

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

// Mux wraps the chi router to provide HTTP request routing.
type Mux struct {
	chi.Router
}

// NewChiMux creates a new chi-based HTTP router with recovery middleware and a health endpoint.
func NewChiMux() *Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/health", Heartbeat("/health"))

	return &Mux{
		Router: r,
	}
}

// Heartbeat creates an HTTP handler that responds to GET or HEAD requests at the specified endpoint.
func Heartbeat(endpoint string) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if (r.Method == "GET" || r.Method == "HEAD") && strings.EqualFold(r.URL.Path, endpoint) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("."))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}

	return fn
}

// Handle wraps chi's Mount to attach another http.Handler along ./pattern/*
func (m *Mux) Handle(pattern string, handler http.Handler) {
	m.Router.Mount(pattern, handler)
}
