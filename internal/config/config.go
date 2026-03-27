package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all CLI configuration contexts.
type Config struct {
	CurrentContext string    `yaml:"current_context"`
	Contexts       []Context `yaml:"contexts"`
}

// Context holds credentials and tokens for a named configuration context.
type Context struct {
	Name             string     `yaml:"name"`
	ClientID         string     `yaml:"client_id"`
	ClientSecret     string     `yaml:"client_secret"`
	AccessToken      string     `yaml:"access_token"`
	RefreshToken     string     `yaml:"refresh_token"`
	TokenExpiresAt   time.Time  `yaml:"token_expires_at"`
	DefaultProjectID *int32     `yaml:"default_project_id,omitempty"`
}

// DefaultConfigPath returns the default path to the config file: ~/.emma/config.yaml
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".emma/config.yaml"
	}
	return filepath.Join(home, ".emma", "config.yaml")
}

// Load loads the config from the given path, creating a default if not found.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := &Config{}
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

// Save writes cfg to path atomically and sets permissions to 0600.
func Save(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	return os.Chmod(path, 0600)
}

// GetCurrentContext returns the currently active context, or nil if none is set.
func (c *Config) GetCurrentContext() *Context {
	for i := range c.Contexts {
		if c.Contexts[i].Name == c.CurrentContext {
			return &c.Contexts[i]
		}
	}
	return nil
}

// SetContext adds or updates a context in the config.
func (c *Config) SetContext(ctx Context) {
	for i := range c.Contexts {
		if c.Contexts[i].Name == ctx.Name {
			c.Contexts[i] = ctx
			return
		}
	}
	c.Contexts = append(c.Contexts, ctx)
}

// DeleteContext removes a context by name from the config.
func (c *Config) DeleteContext(name string) {
	filtered := c.Contexts[:0]
	for _, ctx := range c.Contexts {
		if ctx.Name != name {
			filtered = append(filtered, ctx)
		}
	}
	c.Contexts = filtered
	if c.CurrentContext == name {
		c.CurrentContext = ""
	}
}
