package cmd

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/emma-community/emma-cli/internal/auth"
	"github.com/emma-community/emma-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func (c *CLI) newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Annotations: map[string]string{
			"skipAuth": "true",
		},
	}
	cmd.AddCommand(c.newAuthLoginCmd())
	cmd.AddCommand(c.newAuthLogoutCmd())
	cmd.AddCommand(c.newAuthWhoamiCmd())
	return cmd
}

func (c *CLI) newAuthLoginCmd() *cobra.Command {
	var clientID string
	var clientSecret string
	var contextName string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to the emma platform",
		Annotations: map[string]string{
			"skipAuth": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if clientID == "" {
				return fmt.Errorf("--client-id is required")
			}

			// Get client secret from env if not provided via flag
			if clientSecret == "" {
				clientSecret = os.Getenv("EMMA_CLIENT_SECRET")
			}

			// Prompt interactively if still empty and stdin is a TTY
			if clientSecret == "" {
				if term.IsTerminal(int(syscall.Stdin)) {
					fmt.Fprint(c.Err, "Client Secret: ")
					secretBytes, err := term.ReadPassword(int(syscall.Stdin))
					fmt.Fprintln(c.Err)
					if err != nil {
						return fmt.Errorf("reading client secret: %w", err)
					}
					clientSecret = string(secretBytes)
				} else {
					return fmt.Errorf("--client-secret is required (or set EMMA_CLIENT_SECRET)")
				}
			}

			// Warn if already logged in for this context
			existing := c.Cfg.GetCurrentContext()
			if existing != nil && existing.Name == contextName && existing.AccessToken != "" {
				fmt.Fprintf(c.Err, "Warning: already logged in as context %q — proceeding with re-login.\n", contextName)
			}

			// Issue token
			token, err := auth.IssueToken(clientID, clientSecret)
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			// Build context
			ctx := config.Context{
				Name:         contextName,
				ClientID:     clientID,
				ClientSecret: clientSecret,
			}
			if token.AccessToken != nil {
				ctx.AccessToken = *token.AccessToken
			}
			if token.RefreshToken != nil {
				ctx.RefreshToken = *token.RefreshToken
			}
			if token.ExpiresIn != nil {
				ctx.TokenExpiresAt = time.Now().Add(time.Duration(*token.ExpiresIn) * time.Second)
			}

			c.Cfg.SetContext(ctx)
			if c.Cfg.CurrentContext == "" {
				c.Cfg.CurrentContext = contextName
			}

			if err := config.Save(c.Cfg, c.CfgPath); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			fmt.Fprintln(c.Out, "Logged in.")
			return nil
		},
	}

	cmd.Flags().StringVar(&clientID, "client-id", "", "Client ID (required)")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "Client Secret (or set EMMA_CLIENT_SECRET env var)")
	cmd.Flags().StringVar(&contextName, "context", "default", "Context name to save credentials to")
	_ = cmd.MarkFlagRequired("client-id")

	return cmd
}

func (c *CLI) newAuthLogoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out from the emma platform",
		Annotations: map[string]string{
			"skipAuth": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := c.Cfg.GetCurrentContext()
			if ctx == nil {
				fmt.Fprintln(c.Out, "Not logged in.")
				return nil
			}

			ctx.AccessToken = ""
			ctx.RefreshToken = ""
			ctx.TokenExpiresAt = time.Time{}
			c.Cfg.SetContext(*ctx)

			if err := config.Save(c.Cfg, c.CfgPath); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			fmt.Fprintln(c.Out, "Logged out.")
			return nil
		},
	}
	return cmd
}

func (c *CLI) newAuthWhoamiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show current auth context",
		Annotations: map[string]string{
			"skipAuth": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := c.Cfg.GetCurrentContext()
			if ctx == nil {
				return fmt.Errorf("not logged in — run 'emma auth login' first")
			}
			fmt.Fprintf(c.Out, "Context: %s\nClient ID: %s\n", ctx.Name, ctx.ClientID)
			return nil
		},
	}
	return cmd
}
