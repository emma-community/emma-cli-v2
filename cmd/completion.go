package cmd

import (
	"github.com/spf13/cobra"
)

func (c *CLI) newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Annotations: map[string]string{
			"skipAuth": "true",
		},
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			switch args[0] {
			case "bash":
				return root.GenBashCompletion(c.Out)
			case "zsh":
				return root.GenZshCompletion(c.Out)
			case "fish":
				return root.GenFishCompletion(c.Out, true)
			case "powershell":
				return root.GenPowerShellCompletion(c.Out)
			}
			return nil
		},
	}
	return cmd
}
