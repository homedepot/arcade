package rancher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type NewTokenRequest struct {
	ResponseType string `json:"responseType"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

var (
	errNotFoundFormat = "error getting token: %s"
	mux               sync.Mutex
	expiration        time.Time
	cachedToken       KubeconfigToken
)

func NewClient() *Client {
	return &Client{
		c: &http.Client{},
	}
}

type Client struct {
	url      string
	username string
	password string
	c        *http.Client
}

func (c *Client) WithURL(url string) {
	c.url = url
}

func (c *Client) WithUsername(username string) {
	c.username = username
}

func (c *Client) WithPassword(password string) {
	c.password = password
}

func (c *Client) WithTransport(transport *http.Transport) {
	c.c.Transport = transport
}

func (c *Client) Token(ctx context.Context) (string, error) {
	mux.Lock()
	defer mux.Unlock()

	if time.Now().In(time.UTC).After(cachedToken.ExpiresAt) || cachedToken.Token == "" {
		k := KubeconfigToken{}

		data := NewTokenRequest{
			ResponseType: "kubeconfig",
			Username:     c.username,
			Password:     c.password,
		}

		b, err := json.Marshal(data)
		if err != nil {
			return "", err
		}

		req, err := http.NewRequest(http.MethodPost, c.url, bytes.NewBuffer(b))
		if err != nil {
			return "", err
		}

		req = req.WithContext(ctx)
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")

		res, err := c.c.Do(req)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusCreated {
			_, _ = io.Copy(ioutil.Discard, res.Body)

			return "", fmt.Errorf(errNotFoundFormat, res.Status)
		}

		err = json.NewDecoder(res.Body).Decode(&k)
		if err != nil {
			return "", err
		}

		cachedToken = k
	}

	return cachedToken.Token, nil
}
