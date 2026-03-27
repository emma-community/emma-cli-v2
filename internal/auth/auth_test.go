package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	emma "github.com/emma-community/emma-go-sdk"
	"github.com/emma-community/emma-cli/internal/config"
)

// tokenResponse is a minimal token response for test servers.
// Uses camelCase to match the SDK's Token struct JSON tags.
type tokenResponse struct {
	AccessToken      string `json:"accessToken"`
	RefreshToken     string `json:"refreshToken"`
	ExpiresIn        int32  `json:"expiresIn"`
	RefreshExpiresIn int32  `json:"refreshExpiresIn"`
	TokenType        string `json:"tokenType"`
}

func TestIssueToken_Success(t *testing.T) {
	resp := tokenResponse{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresIn:    600,
		TokenType:    "Bearer",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The SDK adds /external prefix from default server config, so we accept any path ending in /issue-token
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	// Override the SDK configuration with test server URL including /external prefix
	cfg := emma.NewConfiguration()
	cfg.Servers = emma.ServerConfigurations{
		{URL: srv.URL},
	}
	client := emma.NewAPIClient(cfg)

	creds := emma.NewCredentials("test-client-id", "test-secret")
	token, _, err := client.AuthenticationAPI.IssueToken(context.Background()).Credentials(*creds).Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == nil || token.AccessToken == nil {
		t.Fatalf("expected non-nil token with AccessToken, got: %+v", token)
	}
	if *token.AccessToken != "test-access-token" {
		t.Errorf("expected test-access-token, got %q", *token.AccessToken)
	}
}

func TestIssueToken_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer srv.Close()

	cfg := emma.NewConfiguration()
	cfg.Servers = emma.ServerConfigurations{
		{URL: srv.URL},
	}
	client := emma.NewAPIClient(cfg)
	creds := emma.NewCredentials("bad-id", "bad-secret")
	_, _, err := client.AuthenticationAPI.IssueToken(context.Background()).Credentials(*creds).Execute()
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

func TestRefreshToken_Success(t *testing.T) {
	resp := tokenResponse{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		ExpiresIn:    600,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	cfg := emma.NewConfiguration()
	cfg.Servers = emma.ServerConfigurations{
		{URL: srv.URL},
	}
	client := emma.NewAPIClient(cfg)
	rt := emma.NewRefreshToken("old-refresh-token")
	token, _, err := client.AuthenticationAPI.RefreshToken(context.Background()).RefreshToken(*rt).Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == nil || token.AccessToken == nil {
		t.Fatalf("expected non-nil token with AccessToken, got: %+v", token)
	}
	if *token.AccessToken != "new-access-token" {
		t.Errorf("expected new-access-token, got %q", *token.AccessToken)
	}
}

func TestEnsureFreshToken_ValidToken_NoOp(t *testing.T) {
	cfg := &config.Config{
		CurrentContext: "default",
		Contexts: []config.Context{
			{
				Name:           "default",
				AccessToken:    "valid-token",
				TokenExpiresAt: time.Now().Add(5 * time.Minute),
			},
		},
	}

	// Should not call any API — token is still valid
	err := EnsureFreshToken(cfg, "/dev/null")
	if err != nil {
		t.Fatalf("expected no error for valid token, got: %v", err)
	}
	// Token should be unchanged
	if cfg.Contexts[0].AccessToken != "valid-token" {
		t.Errorf("expected token unchanged, got %q", cfg.Contexts[0].AccessToken)
	}
}

func TestEnsureFreshToken_ExpiredToken_Refreshes(t *testing.T) {
	// Without credentials or refresh token, should fail
	cfg := &config.Config{
		CurrentContext: "default",
		Contexts: []config.Context{
			{
				Name:           "default",
				AccessToken:    "expired-token",
				RefreshToken:   "",  // No refresh token
				ClientID:       "",  // No credentials
				TokenExpiresAt: time.Now().Add(-1 * time.Minute), // expired
			},
		},
	}

	err := EnsureFreshToken(cfg, "/tmp/test-config-ensure.yaml")
	if err == nil {
		t.Error("expected error when token expired and no credentials/refresh token available")
	}
}
