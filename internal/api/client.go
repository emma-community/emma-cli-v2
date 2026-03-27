// Package api provides helpers to create a configured emma API client.
package api

import (
	"context"

	emma "github.com/emma-community/emma-go-sdk"
)

// NewClient creates a new emma APIClient configured with the given bearer token.
func NewClient(token string) *emma.APIClient {
	cfg := emma.NewConfiguration()
	cfg.AddDefaultHeader("Authorization", "Bearer "+token)
	return emma.NewAPIClient(cfg)
}

// ContextWithToken returns a context.Context with the bearer token attached for SDK calls.
func ContextWithToken(token string) context.Context {
	return context.WithValue(context.Background(), emma.ContextAccessToken, token)
}
