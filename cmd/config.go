package cmd

import (
	"fmt"

	"github.com/emma-community/emma-cli/internal/config"
	"github.com/emma-community/emma-cli/internal/output"
	"github.com/spf13/cobra"
)

func (c *CLI) newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage emma CLI configuration",
		Annotations: map[string]string{
			"skipAuth": "true",
		},
	}
	cmd.AddCommand(c.newConfigSetContextCmd())
	cmd.AddCommand(c.newConfigGetContextsCmd())
	cmd.AddCommand(c.newConfigUseContextCmd())
	cmd.AddCommand(c.newConfigDeleteContextCmd())
	return cmd
}

func (c *CLI) newConfigSetContextCmd() *cobra.Command {
	var clientID string
	var clientSecret string
	var projectID int32
	var hasProjectID bool

	cmd := &cobra.Command{
		Use:   "set-context <name>",
		Short: "Create or update a named context",
		Annotations: map[string]string{
			"skipAuth": "true",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Load existing or build new
			var ctx config.Context
			existing := c.Cfg.GetCurrentContext()
			for _, cx := range c.Cfg.Contexts {
				if cx.Name == name {
					ctx = cx
					break
				}
			}
			_ = existing

			ctx.Name = name
			if clientID != "" {
				ctx.ClientID = clientID
			}
			if clientSecret != "" {
				ctx.ClientSecret = clientSecret
			}
			if hasProjectID {
				ctx.DefaultProjectID = &projectID
			}

			c.Cfg.SetContext(ctx)
			if err := config.Save(c.Cfg, c.CfgPath); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
			fmt.Fprintf(c.Out, "Context %q saved.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&clientID, "client-id", "", "Client ID")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "Client Secret")
	cmd.Flags().Int32Var(&projectID, "project-id", 0, "Default project ID")
	cmd.Flags().BoolVar(&hasProjectID, "set-project-id", false, "Set the project-id flag")
	_ = hasProjectID

	// Use a pre-run to detect if project-id was explicitly set
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		hasProjectID = cmd.Flags().Changed("project-id")
		return nil
	}

	return cmd
}

func (c *CLI) newConfigGetContextsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-contexts",
		Short: "List all configured contexts",
		Annotations: map[string]string{
			"skipAuth": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			headers := []string{"NAME", "CLIENT-ID", "CURRENT"}
			rows := make([][]string, 0, len(c.Cfg.Contexts))
			for _, ctx := range c.Cfg.Contexts {
				current := ""
				if ctx.Name == c.Cfg.CurrentContext {
					current = "*"
				}
				rows = append(rows, []string{ctx.Name, ctx.ClientID, current})
			}
			return output.Render(c.Out, c.OutputFmt, c.NoColor, c.Cfg.Contexts, output.TableView{
				Headers: headers,
				Rows:    rows,
			})
		},
	}
	return cmd
}

func (c *CLI) newConfigUseContextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use-context <name>",
		Short: "Switch to a named context",
		Annotations: map[string]string{
			"skipAuth": "true",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Verify the context exists
			found := false
			for _, ctx := range c.Cfg.Contexts {
				if ctx.Name == name {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("context %q not found", name)
			}

			c.Cfg.CurrentContext = name
			if err := config.Save(c.Cfg, c.CfgPath); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
			fmt.Fprintf(c.Out, "Switched to context %q.\n", name)
			return nil
		},
	}
	return cmd
}

func (c *CLI) newConfigDeleteContextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-context <name>",
		Short: "Delete a named context",
		Annotations: map[string]string{
			"skipAuth": "true",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			c.Cfg.DeleteContext(name)
			if err := config.Save(c.Cfg, c.CfgPath); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
			fmt.Fprintf(c.Out, "Context %q deleted.\n", name)
			return nil
		},
	}
	return cmd
}
