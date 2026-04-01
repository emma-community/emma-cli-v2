package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// WatchInterval is the default interval between watch refreshes.
const WatchInterval = 5 * time.Second

// addWatchFlag adds the --watch flag to a list command.
func addWatchFlag(cmd *cobra.Command, interval *int) {
	cmd.Flags().IntVar(interval, "watch", 0, "Watch mode: refresh every N seconds (0 = disabled)")
}

// runWatch clears the screen and calls fn repeatedly until context is cancelled.
// fn should render output to the CLI's writer.
func runWatch(ctx context.Context, intervalSec int, fn func() error) error {
	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer ticker.Stop()

	for {
		// Clear screen
		fmt.Fprint(os.Stderr, "\033[2J\033[H")

		if err := fn(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		fmt.Fprintf(os.Stderr, "\nRefreshing every %ds... (Ctrl+C to stop)", intervalSec)

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// continue
		}
	}
}
