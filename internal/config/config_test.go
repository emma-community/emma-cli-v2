package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_FileNotFound(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if len(cfg.Contexts) != 0 {
		t.Errorf("expected empty contexts, got %d", len(cfg.Contexts))
	}
}

func TestLoad_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := `current_context: default
contexts:
  - name: default
    client_id: my-client-id
    client_secret: my-secret
    access_token: tok123
    refresh_token: ref456
    token_expires_at: "2024-01-01T00:00:00Z"
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CurrentContext != "default" {
		t.Errorf("expected current_context=default, got %q", cfg.CurrentContext)
	}
	if len(cfg.Contexts) != 1 {
		t.Fatalf("expected 1 context, got %d", len(cfg.Contexts))
	}
	ctx := cfg.Contexts[0]
	if ctx.ClientID != "my-client-id" {
		t.Errorf("expected client_id=my-client-id, got %q", ctx.ClientID)
	}
	if ctx.AccessToken != "tok123" {
		t.Errorf("expected access_token=tok123, got %q", ctx.AccessToken)
	}
}

func TestLoad_CorruptedYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(path, []byte("{{{invalid yaml"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for corrupted YAML, got nil")
	}
}

func TestSave_Atomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{
		CurrentContext: "test",
		Contexts: []Context{
			{
				Name:     "test",
				ClientID: "cid",
			},
		},
	}

	if err := Save(cfg, path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Tmp file should not exist after save
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error("expected .tmp file to be gone after atomic save")
	}

	// File should exist
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected config file to exist: %v", err)
	}

	// Reload and verify
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}
	if loaded.CurrentContext != "test" {
		t.Errorf("expected current_context=test, got %q", loaded.CurrentContext)
	}
}

func TestSave_Permissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{}
	if err := Save(cfg, path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected permissions 0600, got %o", info.Mode().Perm())
	}
}

func TestGetCurrentContext(t *testing.T) {
	cfg := &Config{
		CurrentContext: "prod",
		Contexts: []Context{
			{Name: "dev", ClientID: "dev-id"},
			{Name: "prod", ClientID: "prod-id"},
		},
	}

	ctx := cfg.GetCurrentContext()
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
	if ctx.ClientID != "prod-id" {
		t.Errorf("expected prod-id, got %q", ctx.ClientID)
	}
}

func TestSetContext(t *testing.T) {
	cfg := &Config{}
	ctx := Context{Name: "new", ClientID: "new-id"}
	cfg.SetContext(ctx)
	if len(cfg.Contexts) != 1 {
		t.Fatalf("expected 1 context, got %d", len(cfg.Contexts))
	}

	// Update existing
	ctx2 := Context{Name: "new", ClientID: "updated-id"}
	cfg.SetContext(ctx2)
	if len(cfg.Contexts) != 1 {
		t.Fatalf("expected still 1 context after update, got %d", len(cfg.Contexts))
	}
	if cfg.Contexts[0].ClientID != "updated-id" {
		t.Errorf("expected updated-id, got %q", cfg.Contexts[0].ClientID)
	}
}

func TestDeleteContext(t *testing.T) {
	cfg := &Config{
		CurrentContext: "ctx1",
		Contexts: []Context{
			{Name: "ctx1"},
			{Name: "ctx2"},
		},
	}
	cfg.DeleteContext("ctx1")
	if len(cfg.Contexts) != 1 {
		t.Fatalf("expected 1 context after delete, got %d", len(cfg.Contexts))
	}
	if cfg.CurrentContext != "" {
		t.Errorf("expected CurrentContext to be cleared, got %q", cfg.CurrentContext)
	}
}

func TestTokenExpiresAt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	expiry := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	cfg := &Config{
		CurrentContext: "default",
		Contexts: []Context{
			{
				Name:           "default",
				TokenExpiresAt: expiry,
			},
		},
	}

	if err := Save(cfg, path); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	got := loaded.Contexts[0].TokenExpiresAt
	if !got.Equal(expiry) {
		t.Errorf("expected %v, got %v", expiry, got)
	}
}
