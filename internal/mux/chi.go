package mux

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Mux struct {
	chi.Router
}

func NewChiMux(router *chi.Mux) *Mux {
	return &Mux{
		Router: router,
	}
}

// Handle wraps chi's Mount to attach another http.Handler along ./pattern/*
func (m *Mux) Handle(pattern string, handler http.Handler) {
	m.Router.Mount(pattern, handler)
}
