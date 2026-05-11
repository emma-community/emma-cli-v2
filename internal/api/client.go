// Package api provides helpers to create a configured emma API client.
package api

import (
	"context"
	"os"

	emma "github.com/emma-community/emma-go-sdk"
)

func newConfiguration() *emma.Configuration {
	cfg := emma.NewConfiguration()
	if url := os.Getenv("EMMA_API_URL"); url != "" {
		cfg.Servers = emma.ServerConfigurations{{URL: url}}
	}
	return cfg
}

// NewClient creates a new emma APIClient configured with the given bearer token.
func NewClient(token string) *emma.APIClient {
	cfg := newConfiguration()
	cfg.AddDefaultHeader("Authorization", "Bearer "+token)
	return emma.NewAPIClient(cfg)
}

// NewAnonymousClient creates a new emma APIClient without authentication.
func NewAnonymousClient() *emma.APIClient {
	return emma.NewAPIClient(newConfiguration())
}

// ContextWithToken returns a context.Context with the bearer token attached for SDK calls.
func ContextWithToken(token string) context.Context {
	return context.WithValue(context.Background(), emma.ContextAccessToken, token)
}
