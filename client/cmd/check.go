package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCheckCommand(opts *options) *cobra.Command {
	var maxRetry, interval, sleep int
	var quiet bool
	command := &cobra.Command{
		Use: "check [SERVICE]", Short: "Monitor registry membership until failure or cancellation", Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if maxRetry < 1 {
				return fmt.Errorf("--max-retry must be at least 1")
			}
			retryInterval, err := secondsDuration("interval", interval, false)
			if err != nil {
				return err
			}
			pollInterval, err := secondsDuration("sleep", sleep, false)
			if err != nil {
				return err
			}
			name, err := serviceArg(args)
			if err != nil {
				return err
			}
			retry, step := 0, uint64(0)
			for {
				if err := cmd.Context().Err(); err != nil {
					return err
				}
				var body []byte
				var available bool
				if quiet {
					available, err = opts.client.Present(cmd.Context(), name)
				} else {
					body, available, err = opts.client.Status(cmd.Context(), name)
				}
				if err != nil {
					return err
				}
				if !available {
					retry++
					if !quiet {
						if _, err := fmt.Fprintf(cmd.OutOrStdout(), "[*] No data available. Retry %d out of %d\n", retry, maxRetry); err != nil {
							return err
						}
					}
					if retry >= maxRetry {
						return fmt.Errorf("max retry reached without available service data")
					}
					if err := wait(cmd.Context(), retryInterval); err != nil {
						return err
					}
					continue
				}
				retry = 0
				if !quiet {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "[*] Step %d\n\tData: %s", step, body); err != nil {
						return err
					}
				}
				step++
				if err := wait(cmd.Context(), pollInterval); err != nil {
					return err
				}
			}
		},
	}
	command.Flags().IntVarP(&maxRetry, "max-retry", "r", 3, "Fail after this many consecutive absent observations")
	command.Flags().IntVarP(&interval, "interval", "t", 5, "Seconds between absent-service retries (positive)")
	command.Flags().IntVarP(&sleep, "sleep", "s", 5, "Seconds between successful polls (positive)")
	command.Flags().BoolVarP(&quiet, "quiet", "q", false, "Use body-free HEAD probes and suppress normal output (updated server required)")
	return command
}
