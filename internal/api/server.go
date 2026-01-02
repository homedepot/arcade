package api

import (
	"fmt"
	"log"
	"net/http"
)

type (
	config struct {
		port    int
		handler *Handler
	}

	ServerOption func(*config)
)

func NewServer(opts ...ServerOption) *http.Server {
	cfg := &config{}

	for _, opt := range opts {
		opt(cfg)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.Handle("GET /", cfg.handler)

	addr := fmt.Sprintf(":%d", cfg.port)
	log.Printf("listening on %s\n", addr)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

func WithPort(port int) ServerOption {
	return func(c *config) {
		c.port = port
	}
}

func WithHandler(h *Handler) ServerOption {
	return func(c *config) {
		c.handler = h
	}
}
