package arcade_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/homedepot/arcade/pkg"
)

func TestNewDefaultClient(t *testing.T) {
	_ = NewDefaultClient("test-api-key")
}

func TestClient_Token(t *testing.T) {
	t.Run("uri is invalid", func(t *testing.T) {
		client := NewClient("::haha", "test-api-key")
		_, err := client.Token("google")
		assert.Error(t, err)
	})

	t.Run("server is not reachable", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		url := s.URL
		s.Close()

		client := NewClient(url, "test-api-key")
		_, err := client.Token("google")
		assert.Error(t, err)
	})

	t.Run("response is not 2XX", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(s.Close)

		client := NewClient(s.URL, "test-api-key")
		_, err := client.Token("google")

		assert.Error(t, err)
		assert.Equal(t, "error getting token: 500 Internal Server Error", err.Error())
	})

	t.Run("server returns bad data", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(";{["))
		}))
		t.Cleanup(s.Close)

		client := NewClient(s.URL, "test-api-key")
		_, err := client.Token("google")

		assert.Error(t, err)
		assert.Equal(t, "invalid character ';' looking for beginning of value", err.Error())
	})

	t.Run("provider is rancher", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "test-api-key", r.Header.Get("api-key"))
			assert.Equal(t, "/tokens", r.URL.Path)
			assert.Equal(t, "rancher", r.URL.Query().Get("provider"))

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"token": "some.bearer.token"})
		}))
		t.Cleanup(s.Close)

		client := NewClient(s.URL, "test-api-key")
		token, err := client.Token("rancher")

		assert.NoError(t, err)
		assert.Equal(t, "some.bearer.token", token)
	})

	t.Run("it succeeds", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "test-api-key", r.Header.Get("api-key"))
			assert.Equal(t, "/tokens", r.URL.Path)
			assert.Equal(t, "google", r.URL.Query().Get("provider"))

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"token": "some.bearer.token"})
		}))
		t.Cleanup(s.Close)

		client := NewClient(s.URL, "test-api-key")
		token, err := client.Token("google")

		assert.NoError(t, err)
		assert.Equal(t, "some.bearer.token", token)
	})
}
