package provider

import "context"

//go:generate counterfeiter . Client

// Client represents the interface to grab a new token from a provider.
type Client interface {
	Token(context.Context) (string, error)
}
