package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/homedepot/arcade/internal/api/middleware"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyAuth(t *testing.T) {
	type resp struct {
		Error string `json:"error"`
	}

	const (
		apiKeyGood = "secret"
		apiKeyBad  = "nope"
	)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handler := middleware.APIKeyAuth(apiKeyGood)(next)

	t.Run("missing Api-Key header returns 403 json error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/tokens", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var out resp
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
		assert.Equal(t, "bad api key", out.Error)
	})

	t.Run("wrong Api-Key header returns 403 json error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/tokens", nil)
		req.Header.Set("Api-Key", apiKeyBad)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var out resp
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
		assert.Equal(t, "bad api key", out.Error)
	})

	t.Run("correct Api-Key header calls next handler", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/tokens", nil)
		req.Header.Set("Api-Key", apiKeyGood)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "ok", rr.Body.String())
	})
}
