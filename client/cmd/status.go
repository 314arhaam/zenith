package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStatusCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use: "status [SERVICE]", Short: "Query registered services", Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := serviceArg(args)
			if err != nil {
				return err
			}
			body, available, err := opts.client.Status(cmd.Context(), name)
			if err != nil {
				return err
			}
			if name != "" && !available {
				return fmt.Errorf("service not found: %s", name)
			}
			_, err = cmd.OutOrStdout().Write(body)
			return err
		},
	}
}
