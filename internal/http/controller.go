package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/homedepot/arcade/internal/google"
	"github.com/homedepot/arcade/internal/microsoft"
	"github.com/homedepot/arcade/internal/rancher"
)

const (
	ProviderTypeRancher   = "rancher"
	ProviderTypeMicrosoft = "microsoft"
	ProviderTypeGoogle    = "google"
)

// Controller holds clients used to grab tokens.
type Controller struct {
	Tokenizers map[string]Tokenizer
}

// Tokenizer defines the interface for a client that can retrieve a token.
type Tokenizer interface {
	Token(context.Context) (string, error)
}

// Provider defines the token provider configuration.
type Provider struct {
	// General config.
	Type string `json:"type"`
	Name string `json:"name"`
	// Rancher config.
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	RootCA   string `json:"rootCA,omitempty"`
	URL      string `json:"url,omitempty"`
	// Microsoft config.
	ClientID      string `json:"clientId,omitempty"`
	ClientSecret  string `json:"clientSecret,omitempty"`
	Resource      string `json:"resource,omitempty"`
	LoginEndpoint string `json:"loginEndpoint,omitempty"`
}

var (
	defaultConfigDir = "/secret/arcade/providers"
)

// NewDefaultController creates a Controller with the default
// configuration directory.
func NewDefaultController() (Controller, error) {
	return NewController(defaultConfigDir)
}

// NewController creates a Controller, retrieving the token providers'
// configuration files from the given directory.
func NewController(dir string) (Controller, error) {
	controller := Controller{
		Tokenizers: map[string]Tokenizer{},
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return controller, err
	}

	if len(files) == 0 {
		return controller, fmt.Errorf("no token providers found in directory: %s", dir)
	}

	for _, f := range files {
		if !f.IsDir() {
			path := filepath.Join(dir, f.Name())

			// Handle symlinks for ConfigMaps.
			ln, err := filepath.EvalSymlinks(path)
			if err == nil {
				path = ln
			}

			b, err := ioutil.ReadFile(path)
			if err != nil {
				// Just continue if we're not able to read the 'file' as the file might be a symlink to
				// a dir when using kubernetes ConfigMaps, for example:
				//
				// drwxr-xr-x    2 root     root          4096 Oct  8 20:38 ..2020_10_08_20_38_50.434422700
				// lrwxrwxrwx    1 root     root            31 Oct  8 20:38 ..data -> ..2020_10_08_20_38_50.434422700
				continue
			}

			p := Provider{}

			err = json.Unmarshal(b, &p)
			if err != nil {
				return controller, err
			}

			if p.Name == "" {
				return controller, fmt.Errorf("no \"name\" found in token provider config file %s", path)
			}

			for name := range controller.Tokenizers {
				if strings.EqualFold(p.Name, name) {
					return controller, fmt.Errorf("duplicate token provider listed: %s", p.Name)
				}
			}

			t := p.Type
			switch t {
			case ProviderTypeGoogle:
				client := google.NewClient()
				controller.Tokenizers[p.Name] = client
			case ProviderTypeMicrosoft:
				if p.ClientID == "" {
					return controller, fmt.Errorf("microsoft token provider file %s missing required \"clientId\" attribute", p.Name)
				}

				if p.ClientSecret == "" {
					return controller, fmt.Errorf("microsoft token provider file %s missing required \"clientSecret\" attribute", p.Name)
				}

				if p.Resource == "" {
					return controller, fmt.Errorf("microsoft token provider file %s missing required \"resource\" attribute", p.Name)
				}

				if p.LoginEndpoint == "" {
					return controller, fmt.Errorf("microsoft token provider file %s missing required \"loginEndpoint\" attribute", p.Name)
				}

				client := microsoft.NewClient()
				client.WithClientID(p.ClientID)
				client.WithClientSecret(p.ClientSecret)
				client.WithResource(p.Resource)
				client.WithLoginEndpoint(p.LoginEndpoint)
				controller.Tokenizers[p.Name] = client
			case ProviderTypeRancher:
				if p.Username == "" {
					return controller, fmt.Errorf("rancher token provider file %s missing required \"username\" attribute", p.Name)
				}

				if p.Password == "" {
					return controller, fmt.Errorf("rancher token provider file %s missing required \"password\" attribute", p.Name)
				}

				if p.URL == "" {
					return controller, fmt.Errorf("rancher token provider file %s missing required \"url\" attribute", p.Name)
				}

				client := rancher.NewClient()
				// If there's a rootCA, then add to HTTP transport
				if p.RootCA != "" {
					rootCAs, _ := x509.SystemCertPool()
					if rootCAs == nil {
						rootCAs = x509.NewCertPool()
					}

					rootCAs.AppendCertsFromPEM([]byte(p.RootCA))

					t := &http.Transport{
						TLSClientConfig: &tls.Config{
							RootCAs: rootCAs,
						},
					}

					client.WithTransport(t)
				}

				client.WithURL(p.URL)
				client.WithUsername(p.Username)
				client.WithPassword(p.Password)
				controller.Tokenizers[p.Name] = client
			default:
				return controller, fmt.Errorf("unsupported token provider type: %s", p.Type)
			}
		}
	}

	return controller, nil
}
