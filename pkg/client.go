package arcade

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	defaultURL = `http://localhost:1982`
)

//go:generate counterfeiter -o arcadefakes . Client
type Client interface {
	Token(string) (string, error)
}

// NewDefaultClient creates a new instance of client with an API Key
// that calls the default URL endpoint.
func NewDefaultClient(apiKey string) Client {
	return NewClient(defaultURL, apiKey)
}

// NewClient creates a new instance of client with a defined API Key
// and URL endpoint.
func NewClient(url, apiKey string) Client {
	return &client{
		apiKey: apiKey,
		url:    url,
	}
}

type client struct {
	apiKey string
	url    string
}

// Token returns a token for a given provider.
func (c *client) Token(tokenProvider string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, c.url+"/tokens", nil)
	if err != nil {
		return "", err
	}

	q := url.Values{}
	q.Add("provider", tokenProvider)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Api-Key", c.apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 399 {
		return "", fmt.Errorf("error getting token: %s", res.Status)
	}

	var response struct {
		Token string `json:"token"`
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(b, &response)
	if err != nil {
		return "", err
	}

	return response.Token, nil
}
