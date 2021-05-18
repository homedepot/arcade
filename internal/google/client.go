package google

import (
	"context"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	clientScopes = []string{
		"https://www.googleapis.com/auth/cloud-platform",
	}
	mux         sync.Mutex
	expiration  time.Time
	cachedToken *oauth2.Token
)

func NewClient() *Client {
	return &Client{}
}

type Client struct{}

func (*Client) Token(ctx context.Context) (string, error) {
	mux.Lock()
	defer mux.Unlock()

	if time.Now().UTC().After(expiration) || cachedToken == nil {
		tokenSource, err := google.DefaultTokenSource(ctx, clientScopes...)
		if err != nil {
			return "", err
		}

		token, err := tokenSource.Token()
		if err != nil {
			return "", err
		}
		// Set the expiration for the Google token to be 90% expiry-threshold.
		// Expiry looks something like '2021-03-26 15:53:24.513497 -0400 EDT m=+3599.302993422'
		expiration = time.Now().UTC().Add((time.Until(token.Expiry) / 10) * 9)
		// Set the cached token.
		cachedToken = token
	}

	return cachedToken.AccessToken, nil
}
