package cmd

// PersistentPreRunE decision tree:
//
//	cmd.Annotations["skipAuth"]=="true" ──► return nil
//	cfg.GetCurrentContext() == nil      ──► error "not logged in"
//	EnsureFreshToken()                  ──► auto-refresh if needed
//	Build APIClient with fresh token    ──► c.Client = NewClient(token)

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/emma-community/emma-cli/internal/api"
	"github.com/emma-community/emma-cli/internal/auth"
	"github.com/spf13/cobra"
)

// NewRootCmd builds and returns the root cobra command.
func (c *CLI) NewRootCmd() *cobra.Command {
	var projectID int32

	rootCmd := &cobra.Command{
		Use:           "emma",
		Short:         "emma cloud platform CLI",
		Long:          "emma is a kubectl-quality CLI for managing resources on the emma cloud platform.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&c.OutputFmt, "output", "o", "table", "Output format: table, json, yaml")
	rootCmd.PersistentFlags().BoolVar(&c.Debug, "debug", false, "Enable debug output")
	rootCmd.PersistentFlags().BoolVar(&c.NoColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().Int32Var(&projectID, "project-id", 0, "Project ID (overrides context default)")

	// Bind projectID to CLI struct after parse
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Handle --no-color and NO_COLOR env var
		if os.Getenv("NO_COLOR") != "" {
			c.NoColor = true
		}

		// Handle project-id flag
		if f := cmd.Flags().Lookup("project-id"); f == nil {
			f = cmd.InheritedFlags().Lookup("project-id")
			if f != nil && f.Changed {
				c.ProjectID = &projectID
			}
		}
		if rootCmd.PersistentFlags().Changed("project-id") {
			c.ProjectID = &projectID
		}

		// Commands with skipAuth annotation bypass auth entirely
		if cmd.Annotations["skipAuth"] == "true" {
			return nil
		}

		// Check for current context
		ctx := c.Cfg.GetCurrentContext()
		if ctx == nil {
			return fmt.Errorf("not logged in — run 'emma auth login' first")
		}

		// Auto-refresh token if needed
		if err := auth.EnsureFreshToken(c.Cfg, c.CfgPath); err != nil {
			return fmt.Errorf("refreshing token: %w", err)
		}

		// Re-fetch context after potential refresh
		ctx = c.Cfg.GetCurrentContext()
		c.Client = api.NewClient(ctx.AccessToken)

		// Wire up debug HTTP logging
		if c.Debug {
			c.Client.GetConfig().HTTPClient = &http.Client{
				Transport: &debugTransport{base: http.DefaultTransport},
			}
			// Re-add auth header since we replaced the client
			c.Client.GetConfig().AddDefaultHeader("Authorization", "Bearer "+ctx.AccessToken)
		}

		return nil
	}

	// Register all subcommands
	rootCmd.AddCommand(c.newAuthCmd())
	rootCmd.AddCommand(c.newVMCmd())
	rootCmd.AddCommand(c.newSpotCmd())
	rootCmd.AddCommand(c.newK8sCmd())
	rootCmd.AddCommand(c.newVolumeCmd())
	rootCmd.AddCommand(c.newSGCmd())
	rootCmd.AddCommand(c.newSSHKeyCmd())
	rootCmd.AddCommand(c.newSubnetCmd())
	rootCmd.AddCommand(c.newProviderCmd())
	rootCmd.AddCommand(c.newMCNCmd())
	rootCmd.AddCommand(c.newWorkflowCmd())
	rootCmd.AddCommand(c.newConfigCmd())
	rootCmd.AddCommand(c.newCompletionCmd())
	rootCmd.AddCommand(c.newVersionCmd())

	return rootCmd
}

// debugTransport wraps an http.RoundTripper and logs request/response details.
type debugTransport struct {
	base http.RoundTripper
}

func (t *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	log.Printf("[DEBUG] %s %s", req.Method, req.URL)
	resp, err := t.base.RoundTrip(req)
	elapsed := time.Since(start)
	if err != nil {
		log.Printf("[DEBUG] %s %s -> error: %v (%s)", req.Method, req.URL, err, elapsed)
	} else {
		log.Printf("[DEBUG] %s %s -> %d (%s)", req.Method, req.URL, resp.StatusCode, elapsed)
	}
	return resp, err
}
