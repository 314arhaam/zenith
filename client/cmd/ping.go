package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func newPingCommand(opts *options) *cobra.Command {
	var until int
	command := &cobra.Command{
		Use: "ping", Short: "Check availability, optionally retrying until a deadline", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			duration, err := secondsDuration("until", until, true)
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			if duration > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, duration)
				defer cancel()
			}
			for attempt := 1; ; attempt++ {
				body, err := opts.client.Ping(ctx)
				if err == nil {
					_, err = cmd.OutOrStdout().Write(body)
					return err
				}
				if duration == 0 {
					return err
				}
				if ctx.Err() != nil {
					return fmt.Errorf("ping stopped: %w", ctx.Err())
				}
				if _, writeErr := fmt.Fprintf(cmd.ErrOrStderr(), "attempt %d failed: %v\n", attempt, err); writeErr != nil {
					return writeErr
				}
				if err := wait(ctx, time.Second); err != nil {
					return fmt.Errorf("ping stopped: %w", err)
				}
			}
		},
	}
	command.Flags().IntVarP(&until, "until", "u", 0, "Overall retry deadline in seconds; 0 makes one request")
	return command
}
