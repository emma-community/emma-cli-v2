package poller

import (
	"context"
	"fmt"
	"time"
)

// WaitFor polls check repeatedly with the given interval until it returns done=true,
// the timeout elapses, or the context is cancelled.
func WaitFor(ctx context.Context, interval, timeout time.Duration, check func() (done bool, err error)) error {
	deadline := time.Now().Add(timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		done, err := check()
		if err != nil {
			return fmt.Errorf("polling check failed: %w", err)
		}
		if done {
			return nil
		}

		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("timed out waiting after %s", timeout)
			}
			return ctx.Err()
		case <-ticker.C:
			// continue polling
		}
	}
}
