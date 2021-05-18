package http

import "github.com/homedepot/arcade/pkg/provider"

// Controller holds clients used to grab tokens.
type Controller struct {
	GoogleClient    provider.Client
	MicrosoftClient provider.Client
	RancherClient   provider.Client
}
