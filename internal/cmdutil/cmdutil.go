// Package cmdutil provides shared flag helpers for cobra commands.
package cmdutil

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

// OutputFlag adds the --output flag (default "table") to cmd.
func OutputFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "table", "Output format: table, json, yaml")
}

// ProjectIDFlag adds the --project-id flag to cmd.
func ProjectIDFlag(cmd *cobra.Command) {
	cmd.Flags().Int32("project-id", 0, "Project ID (overrides default from context)")
}

// ResourceIDFlag adds the --id flag to cmd.
func ResourceIDFlag(cmd *cobra.Command, name string) {
	cmd.Flags().Int32("id", 0, name+" ID")
}

// NoColorFlag adds the --no-color flag to cmd.
func NoColorFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("no-color", false, "Disable colored output")
}

// ConfirmDelete prompts the user for confirmation before a destructive action.
// Returns true if the user confirms, false otherwise.
// If yes is true, skips the prompt entirely.
func ConfirmDelete(w io.Writer, r io.Reader, resource string, id any, yes bool) bool {
	if yes {
		return true
	}
	fmt.Fprintf(w, "Delete %s %v? [y/N] ", resource, id)
	var input string
	fmt.Fscan(r, &input)
	return strings.ToLower(strings.TrimSpace(input)) == "y"
}
