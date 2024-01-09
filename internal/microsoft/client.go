package microsoft

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Client

// Client makes a request for a client token.
type Client struct {
	c             *http.Client
	cachedToken   string
	clientID      string
	clientSecret  string
	expiration    time.Time
	loginEndpoint string
	mux           sync.Mutex
	resource      string
	timeout       time.Duration
}

// NewClient returns an implementation of Client using a default http client.
func NewClient() *Client {
	return &Client{
		c:          http.DefaultClient,
		expiration: time.Time{}, // Reset expiration for a new client instance.
	}
}

type token struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
	ExtExpiresIn string `json:"ext_expires_in"`
	ExpiresOn    string `json:"expires_on"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	AccessToken  string `json:"access_token"`
}

type errorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorCodes       []int  `json:"error_codes"`
	Timestamp        string `json:"timestamp"`
	TraceID          string `json:"trace_id"`
	CorrelationID    string `json:"correlation_id"`
	ErrorURI         string `json:"error_uri"`
}

// Token returns a cached token if it has not expired, otherwise it
// retrieves a new access token and sets the cached token.
func (c *Client) Token(ctx context.Context) (string, error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	// If the cached token has not expired just return it.
	if time.Now().In(time.UTC).Before(c.expiration) {
		return c.cachedToken, nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("resource", c.resource)
	// Configure request to time out.
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	// Create request and URL encode the form.
	r, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		c.loginEndpoint,
		strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("microsoft: error making request: %w", err)
	}

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.c.Do(r)
	if err != nil {
		return "", fmt.Errorf("microsoft: error doing request for new token: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 399 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return "", fmt.Errorf("microsoft: error getting token: %s", res.Status)
		}

		var e errorResponse

		err = json.Unmarshal(body, &e)
		if err != nil {
			return "", fmt.Errorf("microsoft: error getting token: %s", res.Status)
		}

		return "", fmt.Errorf("microsoft: error getting token: %s", e.ErrorDescription)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("microsoft: error reading body: %w", err)
	}

	var t token

	err = json.Unmarshal(body, &t)
	if err != nil {
		return "", fmt.Errorf("microsoft: error unmarshaling body: %w", err)
	}

	c.cachedToken = t.AccessToken

	expiresIn, err := strconv.Atoi(t.ExpiresIn)
	if err != nil {
		return "", fmt.Errorf("microsoft: error converting expiresIn field for token: %s", err)
	}
	// Set expiration to 90% of expires in time.
	c.expiration = time.Now().In(time.UTC).Add(time.Second * time.Duration((expiresIn/10)*9))

	return c.cachedToken, nil
}

// WithClientID sets the client ID.
func (c *Client) WithClientID(clientID string) {
	c.clientID = clientID
}

// WithClientSecret sets the client secret.
func (c *Client) WithClientSecret(clientSecret string) {
	c.clientSecret = clientSecret
}

// WithLoginEndpoint sets the login endpoint, for example
// 'https://login.microsoftonline.com/someone.onmicrosoft.com/oauth2/token'.
func (c *Client) WithLoginEndpoint(loginEndpoint string) {
	c.loginEndpoint = loginEndpoint
}

// WithResource sets the resource, for example https://graph.microsoft.com.
func (c *Client) WithResource(resource string) {
	c.resource = resource
}

// WithTimeout sets the timeout on the http request to retrieve the token.
func (c *Client) WithTimeout(timeout time.Duration) {
	c.timeout = timeout
}
