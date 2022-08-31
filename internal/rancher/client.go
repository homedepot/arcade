package rancher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
)

func NewClient() *Client {
	return &Client{
		c: &http.Client{},
	}
}

func (c *Client) tokenExpired() bool {
	tokenExpired := c.shortExpiration > 0 && int(time.Since(c.cachedToken.Created.In(time.UTC)).Seconds()) > c.shortExpiration

	expiredAtString := c.cachedToken.ExpiresAt
	if expiredAtString == "" {

		return time.Now().In(time.UTC).After(time.UnixMilli(c.cachedToken.CreatedTS+int64(c.cachedToken.TTL))) || c.cachedToken.Token == "" || tokenExpired

	} else {
		expiredAt, _ := time.Parse(time.RFC3339, expiredAtString)

		return time.Now().In(time.UTC).After(expiredAt) || c.cachedToken.Token == "" || tokenExpired

	}

}

type Client struct {
	c               *http.Client
	cachedToken     KubeconfigToken
	mux             sync.Mutex
	password        string
	shortExpiration int // seconds for the expiration
	timeout         time.Duration
	url             string
	username        string
}

func (c *Client) Token(ctx context.Context) (string, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.tokenExpired() {
		k := KubeconfigToken{}

		data := NewTokenRequest{
			ResponseType: "json",
			Username:     c.username,
			Password:     c.password,
		}

		b, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		// Configure request to time out.
		ctx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()
		// Create the request.
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewBuffer(b))
		if err != nil {
			return "", err
		}

		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")

		res, err := c.c.Do(req)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusCreated {
			body, _ := ioutil.ReadAll(res.Body) // Probably this error isn't important
			log.Printf("Error Body: %s, %s\n", string(body), res.Status)
			// _, _ = io.Copy(ioutil.Discard, res.Body)

			return "", fmt.Errorf(errNotFoundFormat, res.Status)
		}

		err = json.NewDecoder(res.Body).Decode(&k)
		if err != nil {
			return "", err
		}

		c.cachedToken = k
	}

	return c.cachedToken.Token, nil
}

// WithPassword sets the password.
func (c *Client) WithPassword(password string) {
	c.password = password
}

// WithTransport sets the http transport.
func (c *Client) WithTransport(transport *http.Transport) {
	c.c.Transport = transport
}

// WithTimeout sets the timeout on the http request to retrieve the token.
func (c *Client) WithTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// WithURL sets the URL.
func (c *Client) WithURL(url string) {
	c.url = url
}

// WithUsername sets the username.
func (c *Client) WithUsername(username string) {
	c.username = username
}

// WithShortExpiration sets the expiration time used for requesting a fresh token.
func (c *Client) WithShortExpiration(shortExpiration int) {
	c.shortExpiration = shortExpiration
}
