package vaultk8s

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/hashicorp/vault/api"
	"github.com/homedepot/arcade/pkg/provider"
)

const (
	ProviderTypeVaultK8s = "vault-k8s"
	lifecycleWithDash    = 3 // "-XX" = dash + 2-char lifecycle code
	prefixToStrip        = 4 // "-XX-" = dash + 2-char lifecycle code + dash
)

func NewClient() *Client {
	return &Client{
		c: &http.Client{},
	}
}

type Client struct {
	c           *http.Client
	mux         sync.Mutex
	password    string
	url         string
	vaultClient *api.Client
}

func (c *Client) Token(ctx context.Context) (string, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	// Configure Vault client
	config := &api.Config{
		Address:    c.url,
		HttpClient: c.c,
	}

	client, err := api.NewClient(config)
	if err != nil {
		return "", fmt.Errorf("error creating vault client: %w", err)
	}

	c.vaultClient = client

	c.vaultClient.SetToken(c.password)

	var cluster_name string

	var ok bool

	if val := ctx.Value(provider.ProviderKey); val != nil {
		if cluster_name, ok = val.(string); !ok {
			return "", fmt.Errorf("invalid cluster name in context")
		}
	} else {
		return "", fmt.Errorf("cluster name not found in context")
	}

	// If the clustername starts with "vault-k8s-" followed by a two character lifecycle code then remove that prefix to obtain the cluster name
	if len(cluster_name) >= ( len(ProviderTypeVaultK8s) + lifecycleWithDash ) {
		cluster_name = cluster_name[len(ProviderTypeVaultK8s)+prefixToStrip:]
	} else {
		return "", fmt.Errorf("invalid cluster name format")
	}

	vault_pattern := os.Getenv("VAULT_K8S_PATH_PATTERN")
	if vault_pattern == "" {
		vault_pattern = "secret/data/[CLUSTER]/kubeconfig"
	}

	vault_path := strings.ReplaceAll(vault_pattern, "[CLUSTER]", cluster_name)

	vault_uri, err := url.Parse(vault_path)
	if err != nil {
		return "", fmt.Errorf("error parsing vault url: %w", err)
	}

	// Read kubeconfig from Vault
	secret, err := c.vaultClient.Logical().Read(vault_uri.String())
	if err != nil {
		return "", fmt.Errorf("error reading kubeconfig from vault: %w", err)
	}

	if secret == nil {
		return "", fmt.Errorf("secret not found at %s", vault_uri.String())
	}

	jsonBytes, err := json.Marshal(secret.Data)
	if err != nil {
		return "", fmt.Errorf("error marshalling secret data: %w", err)
	}

	var kubeconfigToken KubeconfigToken

	err = json.Unmarshal(jsonBytes, &kubeconfigToken)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling secret data: %w", err)
	}

	if len(kubeconfigToken.Data.Users) == 0 {
		return "", fmt.Errorf("no users found in kubeconfig token")
	}

	return kubeconfigToken.Data.Users[0].User.Token, nil
}

func (c *Client) WithPassword(password string) *Client {
	c.password = password
	return c
}

func (c *Client) WithURL(url string) *Client {
	c.url = url
	return c
}
