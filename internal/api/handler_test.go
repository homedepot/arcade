package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/homedepot/arcade/internal/api"
	"github.com/stretchr/testify/assert"
)

type Tokens struct {
	Token string `json:"token"`
	Error string `json:"error"`
}

type fakeTokenizer struct {
	token string
	err   error
}

func (f *fakeTokenizer) Token(context.Context) (string, error) {
	return f.token, f.err
}

func TestGetToken(t *testing.T) {
	newHandler := func(t *testing.T) (*api.Handler, map[string]*fakeTokenizer) {
		t.Helper()

		fakes := map[string]*fakeTokenizer{
			"google":    {token: "valid-google-token"},
			"microsoft": {token: "valid-microsoft-token"},
			"rancher":   {token: "valid-rancher-token"},
		}

		tokenizers := map[string]api.Tokenizer{
			"google":    fakes["google"],
			"microsoft": fakes["microsoft"],
			"rancher":   fakes["rancher"],
		}

		h := api.NewHandler(
			api.WithTokenizers(tokenizers),
		)

		return h, fakes
	}

	decode := func(t *testing.T, rr *httptest.ResponseRecorder) Tokens {
		t.Helper()
		var tokens Tokens
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &tokens))
		return tokens
	}

	t.Run("provider is not supported", func(t *testing.T) {
		h, _ := newHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/tokens?provider=fake", nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		tokens := decode(t, rr)
		assert.Equal(t, "unsupported token provider: fake", tokens.Error)
	})

	t.Run("defaults to google when provider not specified", func(t *testing.T) {
		h, _ := newHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/tokens", nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		tokens := decode(t, rr)
		assert.Equal(t, "valid-google-token", tokens.Token)
	})

	t.Run("google token error returns 500", func(t *testing.T) {
		h, fakes := newHandler(t)
		fakes["google"].err = errors.New("error getting token from google")

		req := httptest.NewRequest(http.MethodGet, "/tokens?provider=google", nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		tokens := decode(t, rr)
		assert.Equal(t, "error getting token from google", tokens.Error)
	})

	t.Run("google token success", func(t *testing.T) {
		h, _ := newHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/tokens?provider=google", nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		tokens := decode(t, rr)
		assert.Equal(t, "valid-google-token", tokens.Token)
	})

	t.Run("microsoft not configured returns 400", func(t *testing.T) {
		fakeGoogle := &fakeTokenizer{token: "valid-google-token"}
		fakeRancher := &fakeTokenizer{token: "valid-rancher-token"}

		h := api.NewHandler(
			api.WithTokenizers(map[string]api.Tokenizer{
				"google":  fakeGoogle,
				"rancher": fakeRancher,
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/tokens?provider=microsoft", nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		tokens := decode(t, rr)
		assert.Equal(t, "unsupported token provider: microsoft", tokens.Error)
	})

	t.Run("microsoft token error returns 500", func(t *testing.T) {
		h, fakes := newHandler(t)
		fakes["microsoft"].err = errors.New("error getting token from microsoft")

		req := httptest.NewRequest(http.MethodGet, "/tokens?provider=microsoft", nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		tokens := decode(t, rr)
		assert.Equal(t, "error getting token from microsoft", tokens.Error)
	})

	t.Run("microsoft token success", func(t *testing.T) {
		h, _ := newHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/tokens?provider=microsoft", nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		tokens := decode(t, rr)
		assert.Equal(t, "valid-microsoft-token", tokens.Token)
	})

	t.Run("rancher not configured returns 400", func(t *testing.T) {
		fakeGoogle := &fakeTokenizer{token: "valid-google-token"}
		fakeMicrosoft := &fakeTokenizer{token: "valid-microsoft-token"}

		h := api.NewHandler(
			api.WithTokenizers(map[string]api.Tokenizer{
				"google":    fakeGoogle,
				"microsoft": fakeMicrosoft,
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/tokens?provider=rancher", nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		tokens := decode(t, rr)
		assert.Equal(t, "unsupported token provider: rancher", tokens.Error)
	})

	t.Run("rancher token error returns 500", func(t *testing.T) {
		h, fakes := newHandler(t)
		fakes["rancher"].err = errors.New("error getting token from rancher")

		req := httptest.NewRequest(http.MethodGet, "/tokens?provider=rancher", nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		tokens := decode(t, rr)
		assert.Equal(t, "error getting token from rancher", tokens.Error)
	})

	t.Run("rancher token success", func(t *testing.T) {
		h, _ := newHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/tokens?provider=rancher", nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		tokens := decode(t, rr)
		assert.Equal(t, "valid-rancher-token", tokens.Token)
	})
}
