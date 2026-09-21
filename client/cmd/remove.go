package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRemoveCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use: "remove SERVICE", Short: "Remove a service", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := serviceArg(args)
			if err != nil {
				return err
			}
			if err := opts.client.Remove(cmd.Context(), name); err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Status Code: 204")
			return err
		},
	}
}
