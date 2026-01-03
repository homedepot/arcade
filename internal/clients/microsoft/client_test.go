package microsoft_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/homedepot/arcade/internal/clients/microsoft"
)

var (
	ctx = context.Background()
)

func newClient(url string) *microsoft.Client {
	c := microsoft.NewClient()
	c.WithLoginEndpoint(url)
	c.WithClientID("fake-client-id")
	c.WithClientSecret("fake-client-secret")
	c.WithResource("fake-resource")
	c.WithTimeout(time.Second)
	return c
}

func TestClient_Token(t *testing.T) {
	t.Run("uri is invalid", func(t *testing.T) {
		c := newClient("http://example.invalid")
		c.WithLoginEndpoint("::haha")

		_, err := c.Token(ctx)
		assert.Error(t, err)
		assert.Equal(t, `microsoft: error making request: parse "::haha": missing protocol scheme`, err.Error())
	})

	t.Run("server is not reachable", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		url := s.URL
		s.Close()

		c := newClient(url)
		c.WithLoginEndpoint("::haha")

		_, err := c.Token(ctx)
		assert.Error(t, err)
	})

	t.Run("response is not 2XX", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))

		c := newClient(s.URL)

		_, err := c.Token(ctx)
		assert.Error(t, err)
		assert.Equal(t, "microsoft: error getting token: 500 Internal Server Error", err.Error())
	})

	t.Run("server returns bad data", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(";{["))
		}))

		c := newClient(s.URL)

		_, err := c.Token(ctx)
		assert.Error(t, err)
		assert.Equal(t,
			"microsoft: error unmarshaling body: invalid character ';' looking for beginning of value",
			err.Error(),
		)
	})

	t.Run("server returns descriptive error", func(t *testing.T) {
		res := `{
			"error_description": "Error - requested resource not allowed",
			"error": "invalid_grant"
		}`

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(res))
		}))

		c := newClient(s.URL)

		_, err := c.Token(ctx)
		assert.Error(t, err)
		assert.Equal(t, "microsoft: error getting token: Error - requested resource not allowed", err.Error())
	})

	t.Run("response times out", func(t *testing.T) {
		// Sleep long enough that a very small timeout will fire.
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"access_token":"x"}`))
		}))

		c := newClient(s.URL)
		c.WithTimeout(1 * time.Nanosecond)

		_, err := c.Token(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})

	t.Run("token is cached", func(t *testing.T) {
		res := `{
			"token_type": "Bearer",
			"expires_in": "3599",
			"ext_expires_in": "3599",
			"expires_on": "1621369811",
			"not_before": "1621365911",
			"resource": "https://graph.microsoft.com",
			"access_token": "fake.bearer.token.cached"
		}`

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(res))
		}))

		c := newClient(s.URL)

		_, err := c.Token(ctx)
		assert.NoError(t, err)

		token, err := c.Token(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "fake.bearer.token.cached", token)
	})

	t.Run("server returns a token", func(t *testing.T) {
		res := `{
			"token_type": "Bearer",
			"expires_in": "3599",
			"ext_expires_in": "3599",
			"expires_on": "1621369811",
			"not_before": "1621365911",
			"resource": "https://graph.microsoft.com",
			"access_token": "fake.bearer.token"
		}`

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(res))
		}))

		c := newClient(s.URL)

		token, err := c.Token(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "fake.bearer.token", token)
	})

	t.Run("two clients", func(t *testing.T) {
		res1 := `{
			"token_type": "Bearer",
			"expires_in": "3599",
			"ext_expires_in": "3599",
			"expires_on": "1621369811",
			"not_before": "1621365911",
			"resource": "https://graph.microsoft.com",
			"access_token": "fake.bearer.token"
		}`

		res2 := `{
			"token_type": "Bearer",
			"expires_in": "3599",
			"ext_expires_in": "3599",
			"expires_on": "1621369811",
			"not_before": "1621365911",
			"resource": "https://graph.microsoft.com",
			"access_token": "another.fake.bearer.token"
		}`

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)

			if strings.Contains(string(body), "another-fake-client-id") {
				_, _ = w.Write([]byte(res2))
			} else {
				_, _ = w.Write([]byte(res1))
			}
		}))

		c := newClient(s.URL)
		c2 := newClient(s.URL)
		c2.WithClientID("another-fake-client-id")
		c2.WithClientSecret("another-fake-client-secret")
		c2.WithResource("another-fake-resource")

		token, err := c.Token(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "fake.bearer.token", token)

		token2, err := c2.Token(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "another.fake.bearer.token", token2)
	})
}
