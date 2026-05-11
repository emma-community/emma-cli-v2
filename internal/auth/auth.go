// Package auth provides token management for the emma CLI.
//
// Token lifecycle:
//
//	┌──────────────┐  IssueToken  ┌──────────────┐
//	│   NO TOKEN   │─────────────►│ VALID TOKEN  │◄─ RefreshToken
//	└──────────────┘              └──────┬───────┘
//	                                     │ age > expiry - 30s
//	                                     ▼
//	┌──────────────┐  fails       ┌──────────────┐
//	│   NO TOKEN   │◄─────────────│   EXPIRED    │
//	└──────────────┘              └──────────────┘
package auth

import (
	"context"
	"fmt"
	"time"

	emma "github.com/emma-community/emma-go-sdk"
	"github.com/emma-community/emma-cli/internal/api"
	"github.com/emma-community/emma-cli/internal/config"
)

// IssueToken requests a new access+refresh token pair using clientID and clientSecret.
func IssueToken(clientID, clientSecret string) (*emma.Token, error) {
	client := api.NewAnonymousClient()

	creds := emma.NewCredentials(clientID, clientSecret)
	req := client.AuthenticationAPI.IssueToken(context.Background()).Credentials(*creds)
	token, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("issuing token: %w", err)
	}
	return token, nil
}

// RefreshToken obtains a new access token using the provided refresh token.
func RefreshToken(refreshToken string) (*emma.Token, error) {
	client := api.NewAnonymousClient()

	rt := emma.NewRefreshToken(refreshToken)
	req := client.AuthenticationAPI.RefreshToken(context.Background()).RefreshToken(*rt)
	token, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}
	return token, nil
}

// EnsureFreshToken auto-refreshes the token if it will expire within 30 seconds.
// If refresh fails, it falls back to issuing a new token if credentials are available.
func EnsureFreshToken(cfg *config.Config, cfgPath string) error {
	ctx := cfg.GetCurrentContext()
	if ctx == nil {
		return fmt.Errorf("no current context set")
	}

	// Token is still valid if expiry is more than 30s away
	if time.Now().Add(30 * time.Second).Before(ctx.TokenExpiresAt) {
		return nil
	}

	// Attempt refresh if we have a refresh token
	if ctx.RefreshToken != "" {
		token, err := RefreshToken(ctx.RefreshToken)
		if err == nil {
			applyToken(ctx, token)
			cfg.SetContext(*ctx)
			return config.Save(cfg, cfgPath)
		}
	}

	// Fall back to re-issuing with stored credentials
	if ctx.ClientID != "" && ctx.ClientSecret != "" {
		token, err := IssueToken(ctx.ClientID, ctx.ClientSecret)
		if err != nil {
			return fmt.Errorf("re-issuing token: %w", err)
		}
		applyToken(ctx, token)
		cfg.SetContext(*ctx)
		return config.Save(cfg, cfgPath)
	}

	return fmt.Errorf("token expired and no credentials available to refresh — please run 'emma auth login'")
}

// applyToken copies token fields into the config context.
func applyToken(ctx *config.Context, token *emma.Token) {
	if token.AccessToken != nil {
		ctx.AccessToken = *token.AccessToken
	}
	if token.RefreshToken != nil {
		ctx.RefreshToken = *token.RefreshToken
	}
	if token.ExpiresIn != nil {
		ctx.TokenExpiresAt = time.Now().Add(time.Duration(*token.ExpiresIn) * time.Second)
	}
}
