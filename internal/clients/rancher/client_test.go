package rancher_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/homedepot/arcade/internal/clients/rancher"
	"github.com/stretchr/testify/assert"
)

var (
	ctx = context.Background()
)

const (
	username = "test-user"
	password = "test-pass"
)

func newClient(url string) *rancher.Client {
	c := rancher.NewClient()
	c.WithURL(url)
	c.WithUsername(username)
	c.WithPassword(password)
	c.WithTimeout(time.Second)
	return c
}

func assertLoginJSON(t *testing.T, gotBody []byte, wantUser, wantPass string) {
	t.Helper()

	var got map[string]any
	err := json.Unmarshal(bytes.TrimSpace(gotBody), &got)
	assert.NoError(t, err)

	assert.Equal(t, "json", got["responseType"])
	assert.Equal(t, wantUser, got["username"])
	assert.Equal(t, wantPass, got["password"])
}

func TestClient_Token(t *testing.T) {
	t.Run("uri is invalid", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		c := newClient(s.URL)

		_, err := c.Token(ctx)
		assert.Error(t, err)
	})

	t.Run("server is not reachable", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		s.Close()

		c := newClient(s.URL)

		_, err := c.Token(ctx)
		assert.Error(t, err)
	})

	t.Run("response is not 201", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(""))
		}))

		c := newClient(s.URL)

		_, err := c.Token(ctx)
		assert.Error(t, err)
		assert.Equal(t, "error getting token: 404 Not Found", err.Error())
	})

	t.Run("response is not 201 with response body", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("NOT FOUND"))
		}))

		c := newClient(s.URL)

		_, err := c.Token(ctx)
		assert.Error(t, err)
		assert.Equal(t, "error getting token: 404 Not Found", err.Error())
	})

	t.Run("response is not 201 with body > 100 chars", func(t *testing.T) {
		longBody := "<html>Some really long message!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!</html>"

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(longBody))
		}))

		c := newClient(s.URL)

		_, err := c.Token(ctx)
		assert.Error(t, err)
		assert.Equal(t, "error getting token: 500 Internal Server Error", err.Error())
	})

	t.Run("response is invalid json", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("{;'iuiuiu"))
		}))

		c := newClient(s.URL)

		_, err := c.Token(ctx)
		assert.Error(t, err)
		assert.Equal(t, "invalid character ';' looking for beginning of object key string", err.Error())
	})

	t.Run("response times out", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(payloadKubeconfigToken))
		}))

		c := newClient(s.URL)
		c.WithTimeout(1 * time.Nanosecond)

		_, err := c.Token(ctx)
		assert.Error(t, err)
		assert.True(t, strings.HasSuffix(err.Error(), "context deadline exceeded"), err.Error())
	})

	t.Run("username is set", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(payloadKubeconfigToken))
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assertLoginJSON(t, body, "new-user", "test-pass")
		}))

		c := newClient(s.URL)
		c.WithUsername("new-user")

		token, err := c.Token(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "kubeconfig-u-i76rfanbw5:ltqlpxqz5hh52sxfxfbxxkk6xw7pzkh7d922cww6m9x6fjskskxwl9", token)
	})

	t.Run("password is set", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(payloadKubeconfigToken))
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assertLoginJSON(t, body, "test-user", "new-pass")
		}))

		c := newClient(s.URL)
		c.WithPassword("new-pass")

		token, err := c.Token(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "kubeconfig-u-i76rfanbw5:ltqlpxqz5hh52sxfxfbxxkk6xw7pzkh7d922cww6m9x6fjskskxwl9", token)
	})

	t.Run("transport is set", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(payloadKubeconfigToken))
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assertLoginJSON(t, body, "test-user", "test-pass")
		}))

		tr := &http.Transport{TLSClientConfig: &tls.Config{}}
		c := newClient(s.URL)
		c.WithTransport(tr)

		token, err := c.Token(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "kubeconfig-u-i76rfanbw5:ltqlpxqz5hh52sxfxfbxxkk6xw7pzkh7d922cww6m9x6fjskskxwl9", token)
	})

	t.Run("token is cached", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(payloadKubeconfigTokenCached))
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assertLoginJSON(t, body, "test-user", "test-pass")
		}))

		c := newClient(s.URL)

		token, err := c.Token(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "fake.token.cached", token)

		token2, err2 := c.Token(ctx)
		assert.NoError(t, err2)
		assert.Equal(t, "fake.token.cached", token2)
	})

	t.Run("it succeeds", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(payloadKubeconfigToken))
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assertLoginJSON(t, body, "test-user", "test-pass")
		}))

		c := newClient(s.URL)

		token, err := c.Token(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "kubeconfig-u-i76rfanbw5:ltqlpxqz5hh52sxfxfbxxkk6xw7pzkh7d922cww6m9x6fjskskxwl9", token)
	})

	t.Run("another client", func(t *testing.T) {
		var n int
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			if n == 0 {
				_, _ = w.Write([]byte(payloadKubeconfigToken))
				assertLoginJSON(t, body, username, password)
			} else {
				_, _ = w.Write([]byte(payloadKubeconfigTokenAnother))
				assertLoginJSON(t, body, "another-test-user", "another-test-pass")
			}
			n++
		}))

		c := newClient(s.URL)

		c2 := newClient(s.URL)
		c2.WithUsername("another-test-user")
		c2.WithPassword("another-test-pass")

		token, err := c.Token(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "kubeconfig-u-i76rfanbw5:ltqlpxqz5hh52sxfxfbxxkk6xw7pzkh7d922cww6m9x6fjskskxwl9", token)

		token2, err2 := c2.Token(ctx)
		assert.NoError(t, err2)
		assert.Equal(t, "another.token", token2)
	})

	t.Run("cached token + shortExpiration set + it has passed", func(t *testing.T) {
		var n int
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			if n == 0 {
				_, _ = w.Write([]byte(payloadKubeconfigToken))
			} else {
				_, _ = w.Write([]byte(payloadKubeconfigTokenAnother))
			}
			n++
		}))

		c := newClient(s.URL)
		c.WithShortExpiration(1)

		token, err := c.Token(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "kubeconfig-u-i76rfanbw5:ltqlpxqz5hh52sxfxfbxxkk6xw7pzkh7d922cww6m9x6fjskskxwl9", token)

		time.Sleep(2 * time.Second)

		token2, err2 := c.Token(ctx)
		assert.NoError(t, err2)
		assert.Equal(t, "another.token", token2)
	})

	t.Run("shortExpiration set + not passed + cached token", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(payloadKubeconfigTokenCached))
		}))

		c := newClient(s.URL)
		c.WithShortExpiration(9223372040)

		token, err := c.Token(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "fake.token.cached", token)

		time.Sleep(3 * time.Second)

		token2, err2 := c.Token(ctx)
		assert.NoError(t, err2)
		assert.Equal(t, "fake.token.cached", token2)
	})

	t.Run("shortExpiration set + not passed + no cached token", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(payloadKubeconfigToken))
		}))

		c := newClient(s.URL)
		c.WithShortExpiration(1)

		token, err := c.Token(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "kubeconfig-u-i76rfanbw5:ltqlpxqz5hh52sxfxfbxxkk6xw7pzkh7d922cww6m9x6fjskskxwl9", token)
	})
}
