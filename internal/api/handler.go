package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/homedepot/arcade/internal/api/middleware"
)

const (
	DefaultTimeoutSeconds = 30

	ProviderTypeRancher   = "rancher"
	ProviderTypeMicrosoft = "microsoft"
	ProviderTypeGoogle    = "google"
)

type (
	Handler struct {
		http.Handler

		Tokenizers map[string]Tokenizer
		apiKey     string
	}

	Option func(*Handler)

	Tokenizer interface {
		Token(context.Context) (string, error)
	}
)

func NewHandler(opts ...Option) *Handler {
	h := &Handler{
		Tokenizers: make(map[string]Tokenizer),
	}

	for _, opt := range opts {
		opt(h)
	}

	mux := http.NewServeMux()

	mux.Handle(
		"GET /tokens",
		middleware.APIKeyAuth(h.apiKey)(
			http.HandlerFunc(h.GetToken),
		),
	)

	h.Handler = mux

	return h
}

func WithAPIKey(key string) Option {
	return func(h *Handler) {
		h.apiKey = key
	}
}

func WithTokenizers(t map[string]Tokenizer) Option {
	return func(h *Handler) {
		h.Tokenizers = t
	}
}

// GetToken returns a new access token for a given provider.
// GetToken returns a new access token for a given provider.
func (h *Handler) GetToken(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = ProviderTypeGoogle
	}

	tokenizer, ok := h.Tokenizers[provider]
	if !ok {
		writeError(
			w,
			http.StatusBadRequest,
			fmt.Sprintf("unsupported token provider: %s", provider),
		)

		return
	}

	t, err := tokenizer.Token(context.Background())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"token": t,
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
