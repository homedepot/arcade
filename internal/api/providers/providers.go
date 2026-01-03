//nolint:cyclop
package providers

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/homedepot/arcade/internal/api"
	"github.com/homedepot/arcade/internal/clients/google"
	"github.com/homedepot/arcade/internal/clients/microsoft"
	"github.com/homedepot/arcade/internal/clients/rancher"
)

type Provider struct {
	// General config.
	Type string `json:"type"`
	Name string `json:"name"`

	// Rancher config.
	Username        string `json:"username,omitempty"`
	Password        string `json:"password,omitempty"`
	RootCA          string `json:"rootCA,omitempty"`
	URL             string `json:"url,omitempty"`
	ShortExpiration int    `json:"shortExpiration,omitempty"`

	// Microsoft config.
	ClientID      string `json:"clientId,omitempty"`
	ClientSecret  string `json:"clientSecret,omitempty"`
	Resource      string `json:"resource,omitempty"`
	LoginEndpoint string `json:"loginEndpoint,omitempty"`
}

func LoadTokenizersFromDir(dir string, timeout time.Duration) (map[string]api.Tokenizer, error) { //nolint:gocognit
	tokenizers := make(map[string]api.Tokenizer)

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no token providers found in directory: %s", dir)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		path := filepath.Join(dir, f.Name())

		// Handle symlinks for ConfigMaps.
		if ln, err := filepath.EvalSymlinks(path); err == nil {
			path = ln
		}

		b, err := os.ReadFile(path)
		if err != nil {
			// Can be symlink-to-dir in k8s configmaps; ignore.
			continue
		}

		var p Provider
		if err := json.Unmarshal(b, &p); err != nil {
			return nil, fmt.Errorf("parse provider config %s: %w", path, err)
		}

		if strings.TrimSpace(p.Name) == "" {
			return nil, fmt.Errorf("no \"name\" found in token provider config file %s", path)
		}

		// Ensure unique (case-insensitive).
		for name := range tokenizers {
			if strings.EqualFold(p.Name, name) {
				return nil, fmt.Errorf("duplicate token provider listed: %s", p.Name)
			}
		}

		tok, err := buildTokenizer(p, timeout)
		if err != nil {
			return nil, fmt.Errorf("provider %q: %w", p.Name, err)
		}

		tokenizers[p.Name] = tok
	}

	if len(tokenizers) == 0 {
		return nil, fmt.Errorf("no usable token providers found in directory: %s", dir)
	}

	return tokenizers, nil
}

func buildTokenizer(p Provider, timeout time.Duration) (api.Tokenizer, error) { //nolint:gocognit
	switch p.Type {
	case api.ProviderTypeGoogle:
		return google.NewClient(), nil

	case api.ProviderTypeMicrosoft:
		if p.ClientID == "" {
			return nil, fmt.Errorf("microsoft token provider file %s missing required \"clientId\" attribute", p.Name)
		}

		if p.ClientSecret == "" {
			return nil, fmt.Errorf("microsoft token provider file %s missing required \"clientSecret\" attribute", p.Name)
		}

		if p.Resource == "" {
			return nil, fmt.Errorf("microsoft token provider file %s missing required \"resource\" attribute", p.Name)
		}

		if p.LoginEndpoint == "" {
			return nil, fmt.Errorf("microsoft token provider file %s missing required \"loginEndpoint\" attribute", p.Name)
		}

		client := microsoft.NewClient()
		client.WithClientID(p.ClientID)
		client.WithClientSecret(p.ClientSecret)
		client.WithResource(p.Resource)
		client.WithLoginEndpoint(p.LoginEndpoint)
		client.WithTimeout(timeout)

		return client, nil

	case api.ProviderTypeRancher:
		if p.Username == "" {
			return nil, fmt.Errorf("rancher token provider file %s missing required \"username\" attribute", p.Name)
		}

		if p.Password == "" {
			return nil, fmt.Errorf("rancher token provider file %s missing required \"password\" attribute", p.Name)
		}

		if p.URL == "" {
			return nil, fmt.Errorf("rancher token provider file %s missing required \"url\" attribute", p.Name)
		}

		client := rancher.NewClient()

		if p.RootCA != "" {
			rootCAs, _ := x509.SystemCertPool()
			if rootCAs == nil {
				rootCAs = x509.NewCertPool()
			}

			if ok := rootCAs.AppendCertsFromPEM([]byte(p.RootCA)); !ok {
				return nil, fmt.Errorf("invalid \"rootCA\" PEM")
			}

			client.WithTransport(&http.Transport{
				TLSClientConfig: &tls.Config{RootCAs: rootCAs},
			})
		}

		client.WithURL(p.URL)
		client.WithUsername(p.Username)
		client.WithPassword(p.Password)
		client.WithTimeout(timeout)
		client.WithShortExpiration(p.ShortExpiration)

		return client, nil

	default:
		return nil, fmt.Errorf("unsupported token provider type: %s", p.Type)
	}
}
