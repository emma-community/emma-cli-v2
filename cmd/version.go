package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (c *CLI) newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the emma CLI version",
		Annotations: map[string]string{
			"skipAuth": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(c.Out, "emma version %s (commit: %s, built: %s)\n", c.version, c.commit, c.date)
			return nil
		},
	}
	return cmd
}
